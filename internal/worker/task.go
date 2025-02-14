package worker

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
)

type sTask struct {
	storage   storage.ITask
	graph     dag.IGraph
	kind      string
	workspace string
	scriptDir string
}

func newTask(taskName string) (*sTask, error) {
	startTime := time.Now()
	defer logx.Debugln(taskName, "耗时", time.Since(startTime))

	t := &sTask{
		storage:   storage.Task(taskName),
		workspace: filepath.Join(config.App.WorkSpace(), taskName),
		scriptDir: filepath.Join(config.App.ScriptDir(), taskName),
	}

	var err error
	defer func() {
		if err != nil {
			state := models.StateFailed
			if t.storage.IsDisable() {
				state = models.StateStopped
			}
			_ = t.storage.Update(&models.STaskUpdate{
				Message:  err.Error(),
				State:    models.Pointer(state),
				OldState: models.Pointer(models.StatePending),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
		}
	}()

	// 如果任务被禁用，则更新所有步骤状态并返回错误
	if t.storage.IsDisable() {
		logx.Infoln("the task is disabled, no execution required", taskName)
		for _, sName := range t.storage.StepNameList("") {
			_ = t.storage.Step(sName).Update(&models.SStepUpdate{
				Message:  "the task is disabled, no execution required",
				State:    models.Pointer(models.StateStopped),
				OldState: models.Pointer(models.StatePending),
				STime:    models.Pointer(time.Now()),
				ETime:    models.Pointer(time.Now()),
			})
		}
		err = errors.New("the task is disabled, no execution required")
		return nil, err
	}

	// 获取任务类型
	t.kind, err = t.storage.Kind()
	if err != nil {
		logx.Errorln(t.name(), err)
		return nil, err
	}

	// 构建 DAG 图形
	stepNames := t.storage.StepNameList("")
	stepVertex := make(map[string]*dag.Vertex, len(stepNames))
	stepInfo := make(map[string]storage.IStep, len(stepNames))
	for _, sName := range stepNames {
		stepInfo[sName] = t.storage.Step(sName)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, sName := range stepNames {
		wg.Add(1)
		go func(sName string) {
			defer wg.Done()
			step := stepInfo[sName]
			if step.IsDisable() {
				logx.Infoln("the step is disabled, no execution required", sName)
				_ = step.Update(&models.SStepUpdate{
					Message:  "the step is disabled, no execution required",
					State:    models.Pointer(models.StateStopped),
					OldState: models.Pointer(models.StatePending),
					STime:    models.Pointer(time.Now()),
					ETime:    models.Pointer(time.Now()),
				})
				return
			}
			vertex := dag.NewVertex(sName, newStep(step, t.kind, t.workspace, t.scriptDir))
			mu.Lock()
			stepVertex[sName] = vertex
			mu.Unlock()
		}(sName)
	}
	wg.Wait()

	if len(stepVertex) == 0 {
		err = errors.New("no enabled steps")
		return nil, err
	}

	t.graph = dag.New(taskName)
	defer func() {
		if err != nil {
			_ = t.graph.Kill()
		}
	}()

	// 添加顶点及依赖关系到 DAG 图
	for sName, vertex := range stepVertex {
		vertex, err = t.graph.AddVertex(vertex)
		if err != nil {
			logx.Errorln(t.name(), err)
			return nil, err
		}
		deps := stepInfo[sName].Depend().List()
		var depVertices []*dag.Vertex
		for _, dep := range deps {
			if v, ok := stepVertex[dep]; ok {
				depVertices = append(depVertices, v)
			}
		}
		if err = vertex.WithDeps(depVertices...); err != nil {
			logx.Errorln(t.name(), err)
			return nil, err
		}
	}

	// 校验 DAG 图形
	if err = t.graph.Validator(true); err != nil {
		logx.Errorln(t.name(), err)
		return nil, err
	}

	return t, nil
}

func (t *sTask) name() string {
	return t.graph.Name()
}

func (t *sTask) run() (err error) {
	// 更新任务状态为运行中
	if err = t.storage.Update(&models.STaskUpdate{
		State:    models.Pointer(models.StateRunning),
		OldState: models.Pointer(models.StatePending),
		STime:    models.Pointer(time.Now()),
		Message:  "task is running",
	}); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	res := new(models.STaskUpdate)
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logx.Errorln(r, string(stack))
			err = errors.Errorf("task panic: %v", r)
		}
		// 清理资源
		t.clearDir()
		res.State = models.Pointer(models.StateStopped)
		res.Message = "task has stopped"
		res.ETime = models.Pointer(time.Now())
		res.OldState = models.Pointer(models.StateRunning)
		if err != nil {
			res.State = models.Pointer(models.StateFailed)
			res.Message = err.Error()
		}
		if updErr := t.storage.Update(res); updErr != nil {
			logx.Warnln(t.name(), updErr)
		}
	}()

	timeout, err := t.storage.Timeout()
	if err != nil {
		logx.Errorln(t.name(), err)
		return
	}
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeoutCause(context.Background(), timeout+time.Minute, common.ErrTimeOut)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// 等待 DAG 图恢复执行
	t.graph.WaitResume()

	if err = t.initDir(); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	if err = t.graph.Run(ctx); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	// 策略模式下，获取最后一个非待执行状态的步骤状态
	if t.kind == common.KindStrategy {
		steps := t.storage.StepList("")
		for i := len(steps) - 1; i >= 0; i-- {
			switch *steps[i].State {
			case models.StateFailed:
				err = errors.New(steps[i].Message)
				return
			case models.StateStopped:
				return
			default:
				continue
			}
		}
		logx.Warnln(t.name(), "no steps executed????")
	}

	return
}

func (t *sTask) initDir() error {
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

func (t *sTask) clearDir() {
	if err := os.RemoveAll(t.scriptDir); err != nil {
		logx.Errorln(t.name(), err)
	}
	if err := os.RemoveAll(t.workspace); err != nil {
		logx.Errorln(t.name(), err)
	}
}
