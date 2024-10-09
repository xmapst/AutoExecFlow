package worker

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	_ "github.com/xmapst/AutoExecFlow/internal/plugins"
	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type task struct {
	storage   backend.ITask
	graph     *dag.Graph
	workspace string
	scriptDir string
}

func newTask(taskName string) *task {
	t := &task{
		storage:   storage.Task(taskName),
		workspace: filepath.Join(config.App.WorkSpace(), taskName),
		scriptDir: filepath.Join(config.App.ScriptDir(), taskName),
	}
	// 禁用时直接跳过
	if t.storage.IsDisable() {
		logx.Infoln("the task is disabled, no execution required", taskName)
		for _, sName := range t.storage.StepNameList("") {
			_ = t.storage.Step(sName).Update(&models.StepUpdate{
				Message:  "the task is disabled, no execution required",
				State:    models.Pointer(models.StateStop),
				OldState: models.Pointer(models.StatePending),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
		}
		_ = t.storage.Update(&models.TaskUpdate{
			Message:  "the task is disabled, no execution required",
			State:    models.Pointer(models.StateStop),
			OldState: models.Pointer(models.StatePending),
			STime:    models.Pointer(time.Now()),
			ETime:    models.Pointer(time.Now()),
		})
		return nil
	}

	// 校验dag图形
	// 1. 创建顶点
	var stepVertex = make(map[string]*dag.Vertex)
	var steps = t.storage.StepNameList("")
	for _, sName := range steps {
		// 跳过禁用的步骤
		if t.storage.Step(sName).IsDisable() {
			logx.Infoln("the step is disabled, no execution required", sName)
			_ = t.storage.Step(sName).Update(&models.StepUpdate{
				Message:  "the step is disabled, no execution required",
				State:    models.Pointer(models.StateStop),
				OldState: models.Pointer(models.StatePending),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
			continue
		}
		stepVertex[sName] = dag.NewVertex(sName, newStep(t.storage.Step(sName), t.workspace, t.scriptDir))
	}
	if len(stepVertex) == 0 {
		logx.Infoln("the task has no enabled steps, no execution required", taskName)
		return nil
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
	for sName, vertex := range stepVertex {
		vertex, err = t.graph.AddVertex(vertex)
		if err != nil {
			logx.Errorln(t.name(), err)
			return nil
		}
		err = vertex.WithDeps(func() []*dag.Vertex {
			var stepFns []*dag.Vertex
			for _, dep := range t.storage.Step(sName).Depend().List() {
				_stepFn, _ok := stepVertex[dep]
				if !_ok {
					continue
				}
				stepFns = append(stepFns, _stepFn)
			}
			return stepFns
		}()...)
		if err != nil {
			logx.Errorln(t.name(), err)
			return nil
		}
	}
	// 4. 校验dag图形
	if err = t.graph.Validator(true); err != nil {
		logx.Errorln(t.name(), err)
		return nil
	}
	return t
}

func (t *task) name() string {
	return t.graph.Name()
}

func (t *task) run() (err error) {
	if err = t.storage.Update(&models.TaskUpdate{
		State:    models.Pointer(models.StateRunning),
		OldState: models.Pointer(models.StatePending),
		STime:    models.Pointer(time.Now()),
		Message:  "task is running",
	}); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	var res = new(models.TaskUpdate)
	defer func() {
		recover()
		// 清理
		t.clearDir()
		// 结束时间
		res.ETime = models.Pointer(time.Now())
		res.OldState = models.Pointer(models.StateRunning)
		if err != nil {
			res.State = models.Pointer(models.StateFailed)
			res.Message = err.Error()
		}
		// 更新数据
		if err = t.storage.Update(res); err != nil {
			logx.Warnln(t.name(), err)
		}
	}()

	timeout, err := t.storage.Timeout()
	if err != nil {
		logx.Errorln(err)
		return
	}
	var ctx, cancel = context.WithCancel(context.Background())
	if timeout > 0 {
		ctx, cancel = context.WithTimeoutCause(context.Background(), timeout+1*time.Minute, common.ErrTimeOut)
	}
	defer cancel()

	// 判断当前图形是否挂起
	t.graph.WaitResume()

	if err = t.initDir(); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	res.State = models.Pointer(models.StateStop)
	res.Message = "task has stopped"
	if err = t.graph.Run(ctx); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	for _, sName := range t.storage.StepNameList("") {
		state, _ := t.storage.Step(sName).State()
		if state == models.StateFailed {
			res.State = models.Pointer(models.StateFailed)
			err = errors.New("step " + sName + " is failed")
			return
		}
	}
	return
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
