package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/queue"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var qname = fmt.Sprintf("%s-worker-%s", utils.ServiceName, utils.HostName())

var (
	pool      = tunny.NewCallback(1)
	workQueue = queue.NewInMemoryBroker()
)

func init() {
	workQueue.Subscribe(context.Background(), qname, func(m any) {
		fn, ok := m.(func())
		if !ok {
			return
		}
		pool.Submit(fn)
	})
}

func SetSize(n int) {
	pool.SetSize(n)
}

func GetSize() int {
	return pool.GetSize()
}

func StopWait() {
	logx.Info("Waiting for all tasks to complete")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	workQueue.Shutdown(ctx)
}

func Submit(taskName string) error {
	t := &task{
		storage:   storage.Task(taskName),
		workspace: filepath.Join(config.App.WorkSpace(), taskName),
		scriptDir: filepath.Join(config.App.ScriptDir(), taskName),
	}
	// 禁用时直接跳过
	if t.storage.IsDisable() {
		logx.Infoln("the task is disabled, no execution required", taskName)
		for _, name := range t.storage.StepNameList("") {
			_ = t.storage.Step(name).Update(&models.StepUpdate{
				Message:  "the task is disabled, no execution required",
				State:    models.Pointer(models.StateStop),
				OldState: models.Pointer(models.StatePending),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
		}
		return t.storage.Update(&models.TaskUpdate{
			Message:  "the task is disabled, no execution required",
			State:    models.Pointer(models.StateStop),
			OldState: models.Pointer(models.StatePending),
			STime:    models.Pointer(time.Now()),
			ETime:    models.Pointer(time.Now()),
		})
	}

	// 校验dag图形
	// 1. 创建顶点
	var stepVertex = make(map[string]*dag.Vertex)
	var steps = t.storage.StepNameList("")
	for _, name := range steps {
		// 跳过禁用的步骤
		if t.storage.Step(name).IsDisable() {
			logx.Infoln("the step is disabled, no execution required", name)
			_ = t.storage.Step(name).Update(&models.StepUpdate{
				Message:  "the step is disabled, no execution required",
				State:    models.Pointer(models.StateStop),
				OldState: models.Pointer(models.StatePending),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
			continue
		}
		stepVertex[name] = dag.NewVertex(name, t.newStep(name))
	}
	if len(stepVertex) == 0 {
		return errors.New("no step found")
	}

	// 2. 创建dag图形
	t.graph = dag.New(taskName)
	var err error
	defer func() {
		if err != nil {
			_ = t.graph.Kill()
		}
	}()
	// 3. 创建顶点依赖关系
	for name, vertex := range stepVertex {
		vertex, err = t.graph.AddVertex(vertex)
		if err != nil {
			return err
		}
		err = vertex.WithDeps(func() []*dag.Vertex {
			var stepFns []*dag.Vertex
			for _, dep := range t.storage.Step(name).Depend().List() {
				_stepFn, _ok := stepVertex[dep]
				if !_ok {
					continue
				}
				stepFns = append(stepFns, _stepFn)
			}
			return stepFns
		}()...)
		if err != nil {
			return err
		}
	}
	// 4. 校验dag图形
	if err = t.graph.Validator(true); err != nil {
		return err
	}

	return workQueue.Publish(qname, func() {
		runTask(t)
	})
}

func runTask(t *task) {
	var err error
	if err = t.storage.Update(&models.TaskUpdate{
		State:    models.Pointer(models.StateRunning),
		OldState: models.Pointer(models.StatePending),
		STime:    models.Pointer(time.Now()),
		Message:  "task is running",
	}); err != nil {
		return
	}

	defer func() {
		if err != nil {
			logx.Errorln(t.name(), err)
			_ = t.storage.Update(&models.TaskUpdate{
				State:    models.Pointer(models.StateFailed),
				OldState: models.Pointer(models.StateRunning),
				ETime:    models.Pointer(time.Now()),
				Message:  err.Error(),
			})
		}
	}()

	timeout, err := t.storage.Timeout()
	if err != nil {
		return
	}
	var ctx, cancel = context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeoutCause(context.Background(), timeout+1*time.Minute, common.ErrTimeOut)
	}
	defer cancel()
	res := t.run(ctx)
	if res != nil {
		logx.Infoln(t.name(), res)
	}
	return
}
