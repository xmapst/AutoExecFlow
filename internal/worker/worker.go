package worker

import (
	"context"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/deque"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var (
	// DefaultSize 默认worker数为cpu核心数的两倍
	DefaultSize = runtime.NumCPU() * 2
	pool        = tunny.NewCallback(DefaultSize)
	queue       = deque.New[func()]()
)

func init() {
	go dispatch()
}

func SetSize(n int) {
	pool.SetSize(n)
}

func GetSize() int {
	return pool.GetSize()
}

func GetTotal() int64 {
	return storage.TaskCount()
}

func Running() int64 {
	if pool.QueueLength() > int64(pool.GetSize()) {
		return pool.QueueLength() - 1
	}
	return pool.QueueLength()
}

func Waiting() int64 {
	if pool.QueueLength() > int64(pool.GetSize()) {
		return int64(queue.Len()) + 1
	}
	return int64(queue.Len())
}

func StopWait() {
	logx.Info("Waiting for all tasks to complete")
	for queue.Len() != 0 || pool.QueueLength() != 0 {
		time.Sleep(100 * time.Millisecond)
	}
	logx.Info("All tasks completed, normal end")
}

func dispatch() {
	for {
		if queue.Len() == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		fn, err := queue.PopFront()
		if err != nil {
			logx.Errorln("Failed to pop front from queue:", err)
			continue
		}
		pool.Submit(fn)
	}
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
				State:    models.Pointer(models.Stop),
				OldState: models.Pointer(models.Stop),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
		}
		return t.storage.Update(&models.TaskUpdate{
			Message:  "the task is disabled, no execution required",
			State:    models.Pointer(models.Stop),
			OldState: models.Pointer(models.Stop),
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
				State:    models.Pointer(models.Stop),
				OldState: models.Pointer(models.Stop),
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

	// 3. 创建顶点依赖关系
	for name, vertex := range stepVertex {
		var err error
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
	if err := t.graph.Validator(); err != nil {
		return err
	}

	queue.PushBack(func() {
		if err := t.storage.Update(&models.TaskUpdate{
			State:    models.Pointer(models.Running),
			OldState: models.Pointer(models.Pending),
			STime:    models.Pointer(time.Now()),
			Message:  "task is running",
		}); err != nil {
			return
		}
		var err error
		defer func() {
			if err != nil {
				logx.Errorln(t.name(), err)
				_ = t.storage.Update(&models.TaskUpdate{
					State:    models.Pointer(models.Failed),
					OldState: models.Pointer(models.Running),
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
	})
	return nil
}
