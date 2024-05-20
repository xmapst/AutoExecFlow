package worker

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	_ "github.com/xmapst/osreapi/internal/plugins"
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/logx"
)

type task struct {
	storage   backend.ITask
	graph     *dag.Graph
	workspace string
	scriptDir string
}

func (t *task) name() string {
	return t.graph.Name()
}

func (t *task) newStep(name string) dag.VertexFunc {
	s := &step{
		storage:   t.storage.Step(name),
		wg:        new(sync.WaitGroup),
		workspace: t.workspace,
		scriptDir: t.scriptDir,
		logChan:   make(chan string, 15),
	}
	return s.vertexFunc()
}

func (t *task) run(ctx context.Context) error {
	defer func() {
		err := recover()
		if err != nil {
			logx.Errorln(t.name(), err)
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
		if err := t.storage.Update(res); err != nil {
			logx.Errorln(t.name(), err)
		}
	}()

	if err := t.initDir(); err != nil {
		logx.Errorln(t.name(), err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		return nil
	}

	res.State = models.Pointer(models.Stop)
	res.Message = "task has stopped"
	if err := t.graph.Run(ctx); err != nil {
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		logx.Errorln(t.name(), err)
		return err
	}

	for _, name := range t.storage.StepNameList("") {
		state, _ := t.storage.Step(name).State()
		if state == models.Failed {
			res.State = models.Pointer(models.Failed)
			return errors.New("step " + name + " is failed")
		}
	}
	return nil
}

func (t *task) initDir() error {
	if err := utils.EnsureDirExist(t.workspace); err != nil {
		logx.Errorln(t.name(), t.workspace, t.scriptDir, err)
		return err
	}
	if err := utils.EnsureDirExist(t.scriptDir); err != nil {
		logx.Errorln(t.name(), err)
		return err
	}
	return nil
}

func (t *task) clearDir() {
	if err := os.RemoveAll(t.scriptDir); err != nil {
		logx.Errorln(t.name, err)
	}
	if err := os.RemoveAll(t.workspace); err != nil {
		logx.Errorln(t.name, err)
	}
}
