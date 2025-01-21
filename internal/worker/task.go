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
	graph     *dag.Graph
	workspace string
	scriptDir string
}

func newTask(taskName string) (*sTask, error) {
	sTime := time.Now()
	defer func() {
		logx.Debugln(taskName, "耗时", time.Since(sTime))
	}()
	t := &sTask{
		storage:   storage.Task(taskName),
		workspace: filepath.Join(config.App.WorkSpace(), taskName),
		scriptDir: filepath.Join(config.App.ScriptDir(), taskName),
	}
	var err error
	defer func() {
		if err == nil {
			return
		}
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
	}()
	// 禁用时直接跳过
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

	// 校验dag图形
	// 1. 创建顶点
	var stepVertex = make(map[string]*dag.Vertex)
	var steps = t.storage.StepNameList("")

	// 缓存步骤信息
	stepInfo := make(map[string]storage.IStep)
	for _, sName := range steps {
		stepInfo[sName] = t.storage.Step(sName)
	}

	// 使用并行的方式创建顶点
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, sName := range steps {
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
			vertex := dag.NewVertex(sName, newStep(step, t.workspace, t.scriptDir))

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

	// 2. 创建dag图形
	t.graph = dag.New(taskName)
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
			return nil, err
		}
		deps := stepInfo[sName].Depend().List()
		var stepFns []*dag.Vertex
		for _, dep := range deps {
			if _stepFn, _ok := stepVertex[dep]; _ok {
				stepFns = append(stepFns, _stepFn)
			}
		}
		err = vertex.WithDeps(stepFns...)
		if err != nil {
			logx.Errorln(t.name(), err)
			return nil, err
		}
	}
	// 4. 校验dag图形
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
	if err = t.storage.Update(&models.STaskUpdate{
		State:    models.Pointer(models.StateRunning),
		OldState: models.Pointer(models.StatePending),
		STime:    models.Pointer(time.Now()),
		Message:  "task is running",
	}); err != nil {
		logx.Errorln(t.name(), err)
		return
	}

	var res = new(models.STaskUpdate)
	defer func() {
		if _r := recover(); _r != nil {
			stack := debug.Stack()
			logx.Errorln(_r, string(stack))
			err = errors.Errorf("task is panic, %v", _r)
		}
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

	res.State = models.Pointer(models.StateStopped)
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
		logx.Errorln(t.name, err)
	}
	if err := os.RemoveAll(t.workspace); err != nil {
		logx.Errorln(t.name, err)
	}
}
