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

type Task struct {
	backend.ITask
	graph     *dag.Graph
	workspace string
	scriptDir string
}

func NewTask(taskName string) *Task {
	return &Task{
		ITask:     storage.Task(taskName),
		graph:     dag.New(taskName),
		workspace: filepath.Join(config.App.WorkSpace, taskName),
		scriptDir: filepath.Join(config.App.ScriptDir, taskName),
	}
}

func (t *Task) AddVertex(v *dag.Vertex) (*dag.Vertex, error) {
	return t.graph.AddVertex(v)
}

func (t *Task) Validator() error {
	return t.graph.Validator()
}

func (t *Task) Submit() {
	atomic.AddInt64(&taskTotal, 1)
	queue.PushBack(func() {
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

		taskDetail, err := t.Timeout()
		if err != nil {
			logx.Errorln(err)
			return
		}
		var ctx, cancel = context.WithCancel(context.Background())
		if taskDetail > 0 {
			ctx, cancel = context.WithTimeoutCause(context.Background(), (taskDetail*time.Minute)+1, exec.ErrTimeOut)
		}
		defer cancel()
		res := t.run(ctx)
		logx.Infoln(t.Name(), "end of execution")
		if res != nil {
			logx.Infoln(t.Name(), res)
		}
	})
	return
}

func (t *Task) run(ctx context.Context) error {
	defer func() {
		err := recover()
		if err != nil {
			logx.Errorln(t.Name(), err)
		}
	}()

	// 判断当前图形是否挂起
	t.graph.WaitResume()

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

func (t *Task) initDir() error {
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

func (t *Task) clearDir() {
	if err := os.RemoveAll(t.scriptDir); err != nil {
		logx.Errorln(t.Name(), err)
	}
	if err := os.RemoveAll(t.workspace); err != nil {
		logx.Errorln(t.Name(), err)
	}
}
