package worker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	_ "github.com/xmapst/osreapi/internal/plugins"
	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
)

type task struct {
	backend.ITask
	graph     *dag.Graph
	workspace string
	scriptDir string
}

func Submit(taskName string) (err error) {
	atomic.AddInt64(&taskTotal, 1)
	queue.PushBack(func() {
		var _task = task{
			ITask:     storage.Task(taskName),
			graph:     dag.New(taskName),
			workspace: filepath.Join(config.App.WorkSpace, taskName),
			scriptDir: filepath.Join(config.App.ScriptDir, taskName),
		}
		defer func() {
			if err != nil {
				logx.Errorln(_task.Name(), err)
				_ = _task.Update(&models.TaskUpdate{
					State:    models.Pointer(models.Failed),
					OldState: models.Pointer(models.Running),
					STime:    models.Pointer(time.Now()),
					Message:  err.Error(),
				})
			}
		}()

		var _steps = _task.StepList("")
		// 生成顶点
		var stepFnMap = make(map[string]*dag.Vertex)
		for _, _step := range _steps {
			stepFnMap[_step.Name] = dag.NewVertex(_step.Name, _task.newStepVertexFunc(_task.workspace, _task.scriptDir))
		}

		// 编排步骤: 图中的每个顶点都是一个作业
		for _, _step := range _steps {
			stepFn, ok := stepFnMap[_step.Name]
			if !ok {
				continue
			}
			// 添加顶点以及设置依赖关系
			_vertex, err := _task.graph.AddVertex(stepFn)
			if err != nil {
				return
			}
			_vertex.WithDeps(func() []*dag.Vertex {
				var stepFns []*dag.Vertex
				for _, name := range _task.Step(_step.Name).Depend().List() {
					_stepFn, _ok := stepFnMap[name]
					if !_ok {
						continue
					}
					stepFns = append(stepFns, _stepFn)
				}
				return stepFns
			}()...)
		}
		// 校验dag图形
		if err = _task.graph.Validator(); err != nil {
			logx.Errorln(_task.Name(), err)
			return
		}
		taskDetail, err := _task.Timeout()
		if err != nil {
			logx.Errorln(err)
			return
		}
		var ctx, cancel = context.WithCancel(context.Background())
		if taskDetail > 0 {
			ctx, cancel = context.WithTimeoutCause(context.Background(), (taskDetail*time.Minute)+1, exec.ErrTimeOut)
		}
		defer cancel()
		res := _task.run(ctx)
		logx.Infoln(_task.Name(), "end of execution")
		if res != nil {
			logx.Infoln(_task.Name(), res)
		}
	})
	return
}

func (t *task) run(ctx context.Context) error {
	defer func() {
		err := recover()
		if err != nil {
			logx.Errorln(t.Name(), err)
		}
	}()

	// 判断当前图形是否挂起
	if t.graph.Paused() {
		t.graph.WaitResume()
	}
	if err := t.Update(&models.TaskUpdate{
		State:    models.Pointer(models.Running),
		OldState: models.Pointer(models.Pending),
		STime:    models.Pointer(time.Now()),
		Message:  "task is running",
	}); err != nil {
		logx.Errorln(t.Name(), err)
		return err
	}

	var res = new(models.TaskUpdate)
	defer func() {
		// 清理
		t.clearDir()
		// 结束时间
		res.ETime = models.Pointer(time.Now())
		res.OldState = models.Pointer(models.Running)
		// 更新数据
		if err := t.Update(res); err != nil {
			logx.Errorln(t.Name(), err)
		}
	}()

	if err := t.initDir(); err != nil {
		logx.Errorln(t.Name(), err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		return nil
	}

	res.State = models.Pointer(models.Stop)
	res.Message = "task has stopped"
	if err := t.graph.Run(ctx); err != nil {
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		logx.Errorln(t.Name(), err)
		return err
	}

	for _, _step := range t.StepList("") {
		state, _ := t.Step(_step.Name).GetState()
		if state == models.Failed {
			res.State = models.Pointer(models.Failed)
			return errors.New("step " + _step.Name + " is failed")
		}
	}
	return nil
}

func (t *task) initDir() error {
	if err := utils.EnsureDirExist(t.workspace); err != nil {
		logx.Errorln(t.Name(), t.workspace, t.scriptDir, err)
		return err
	}
	if err := utils.EnsureDirExist(t.scriptDir); err != nil {
		logx.Errorln(t.Name(), err)
		return err
	}
	return nil
}

func (t *task) clearDir() {
	if err := os.RemoveAll(t.scriptDir); err != nil {
		logx.Errorln(t.Name(), err)
	}
	if err := os.RemoveAll(t.workspace); err != nil {
		logx.Errorln(t.Name(), err)
	}
}
