package worker

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sync/atomic"
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
	taskTotal   int64
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
	return atomic.LoadInt64(&taskTotal)
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

func Submit(taskName string) {
	atomic.AddInt64(&taskTotal, 1)
	queue.PushBack(func() {
		t := &task{
			ITask:     storage.Task(taskName),
			graph:     dag.New(taskName),
			workspace: filepath.Join(config.App.WorkSpace, taskName),
			scriptDir: filepath.Join(config.App.ScriptDir, taskName),
		}
		if err := t.Update(&models.TaskUpdate{
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
				logx.Errorln(t.Name(), err)
				_ = t.Update(&models.TaskUpdate{
					State:    models.Pointer(models.Failed),
					OldState: models.Pointer(models.Running),
					ETime:    models.Pointer(time.Now()),
					Message:  err.Error(),
				})
			}
		}()

		// 校验dag图形
		// 1. 创建顶点并入库
		var stepVertex = make(map[string]*dag.Vertex)
		var steps = t.StepList("")
		if len(steps) == 0 {
			err = errors.New("no step found")
			return
		}
		for _, s := range steps {
			stepVertex[s.Name] = dag.NewVertex(s.Name, t.newStep())
		}

		// 2. 创建顶点依赖关系
		for _, s := range steps {
			vertex, ok := stepVertex[s.Name]
			if !ok {
				err = fmt.Errorf("%s vertex does not exist", s.Name)
				return
			}

			vertex, err = t.graph.AddVertex(vertex)
			if err != nil {
				return
			}
			err = vertex.WithDeps(func() []*dag.Vertex {
				var stepFns []*dag.Vertex
				for _, name := range t.Step(s.Name).Depend().List() {
					_stepFn, _ok := stepVertex[name]
					if !_ok {
						continue
					}
					stepFns = append(stepFns, _stepFn)
				}
				return stepFns
			}()...)
			if err != nil {
				return
			}
		}
		// 3. 校验dag图形
		err = t.graph.Validator()
		if err != nil {
			return
		}

		taskDetail, err := t.Timeout()
		if err != nil {
			return
		}
		var ctx, cancel = context.WithCancel(context.Background())
		if taskDetail > 0 {
			ctx, cancel = context.WithTimeoutCause(context.Background(), (taskDetail*time.Minute)+1, exec.ErrTimeOut)
		}
		defer cancel()
		res := t.run(ctx)
		if res != nil {
			logx.Infoln(t.Name(), res)
		}
		return
	})
}
