package worker

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/deque"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
	"github.com/xmapst/osreapi/pkg/tunny"
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
		fn, err := queue.PopFront()
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		pool.Submit(fn)
	}
}

func Submit(taskName string) error {
	t := &task{
		storage:   storage.Task(taskName),
		graph:     dag.New(taskName),
		workspace: filepath.Join(config.App.WorkSpace, taskName),
		scriptDir: filepath.Join(config.App.ScriptDir, taskName),
	}

	// 校验dag图形
	// 1. 创建顶点并入库
	var stepVertex = make(map[string]*dag.Vertex)
	var steps = t.storage.StepNameList("")
	if len(steps) == 0 {
		return errors.New("no step found")
	}
	for _, name := range steps {
		stepVertex[name] = dag.NewVertex(name, t.newStep(name))
	}

	// 2. 创建顶点依赖关系
	for _, name := range steps {
		vertex, ok := stepVertex[name]
		if !ok {
			return fmt.Errorf("%s vertex does not exist", name)
		}
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
	// 3. 校验dag图形
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
			ctx, cancel = context.WithTimeoutCause(context.Background(), timeout+1*time.Minute, exec.ErrTimeOut)
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
