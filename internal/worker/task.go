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

	scriptDir string
	workspace string
	graph     *dag.Graph

	Name    string
	Timeout time.Duration
	Steps   []*Step
}

func (t *Task) SaveEnv(env map[string]string) (err error) {
	if len(env) == 0 {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		storage.Task(t.Name).ClearAll()
	}()
	var envs []*models.Env
	for name, value := range env {
		envs = append(envs, &models.Env{
			Name:  name,
			Value: value,
		})
	}
	if err = storage.Task(t.Name).Env().Create(envs); err != nil {
		return err
	}
	return
}

func Submit(task *Task) (err error) {
	task.ITask = storage.Task(task.Name)
	defer func() {
		if err == nil {
			return
		}
		task.ClearAll()
	}()
	// 检查是否存在相同任务正在运行或者在队列中
	state, err := task.GetState()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	if state == models.Running || state == models.Pending || state == models.Paused {
		return errors.New("task already running, please try again later")
	}
	atomic.AddInt64(&taskTotal, 1)
	task.workspace = filepath.Join(config.App.WorkSpace, task.Name)
	task.scriptDir = filepath.Join(config.App.ScriptDir, task.Name)

	var stepFnMap = make(map[string]*dag.Vertex)
	for _, step := range task.Steps {
		step.IStep = task.Step(step.Name)
		step.workspace = task.workspace
		step.scriptDir = task.scriptDir
		stepFn, err := step.build(task.Env())
		if err != nil {
			logx.Errorln(err)
			return err
		}
		stepFnMap[step.Name] = dag.NewVertex(step.Name, stepFn)
	}

	defer func() {
		if err != nil {
			logx.Errorln(task.Name, task.workspace, task.scriptDir, err)
		}
	}()

	// 编排步骤: 创建一个有向无环图，图中的每个顶点都是一个作业
	task.graph = dag.New(task.Name)
	for _, step := range task.Steps {
		stepFn, ok := stepFnMap[step.Name]
		if !ok {
			continue
		}
		// 添加顶点以及设置依赖关系
		vertex, err := task.graph.AddVertex(stepFn)
		if err != nil {
			logx.Errorln(err)
			return err
		}
		vertex.WithDeps(task.buildDeps(stepFnMap, step)...)
	}
	// 校验dag图形
	if err = task.graph.Validator(); err != nil {
		logx.Errorln(task.Name, task.workspace, task.scriptDir, err)
		return err
	}

	// 插入数据
	if err = task.Create(&models.Task{
		Count:   models.Pointer(len(task.Steps)),
		Timeout: task.Timeout,
		TaskUpdate: models.TaskUpdate{
			Message:  "waiting for dispatch",
			State:    models.Pointer(models.Pending),
			OldState: models.Pointer(models.Pending),
			STime:    models.Pointer(time.Now()),
		},
	}); err != nil {
		logx.Errorln(task.Name, task.workspace, task.scriptDir, err)
		return err
	}
	queue.PushBack(func() {
		var ctx, cancel = context.WithCancel(context.Background())
		if task.Timeout > 0 {
			ctx, cancel = context.WithTimeoutCause(context.Background(), task.Timeout+1*time.Minute, exec.ErrTimeOut)
		}
		defer cancel()
		res := task.run(ctx)
		logx.Infoln(task.Name, task.workspace, task.scriptDir, "end of execution")
		if res != nil {
			logx.Infoln(task.Name, task.workspace, task.scriptDir, res)
		}
	})
	return
}

func (t *Task) buildDeps(stepFnMap map[string]*dag.Vertex, step *Step) []*dag.Vertex {
	var stepFns []*dag.Vertex
	for _, name := range step.Depend().List() {
		_stepFn, _ok := stepFnMap[name]
		if !_ok {
			continue
		}
		stepFns = append(stepFns, _stepFn)
	}
	return stepFns
}

func (t *Task) run(ctx context.Context) error {
	defer func() {
		err := recover()
		if err != nil {
			logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
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
		logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
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
			logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
		}
	}()

	if err := t.initDir(); err != nil {
		logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		return nil
	}

	res.State = models.Pointer(models.Stop)
	res.Message = "task has stopped"
	if err := t.graph.Run(ctx); err != nil {
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
		return err
	}

	for _, step := range t.Steps {
		state, _ := step.GetState()
		if state == models.Failed {
			res.State = models.Pointer(models.Failed)
			return errors.New("step " + step.Name + " is failed")
		}
	}
	return nil
}

func (t *Task) initDir() error {
	if err := utils.EnsureDirExist(t.workspace); err != nil {
		logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
		return err
	}
	if err := utils.EnsureDirExist(t.scriptDir); err != nil {
		logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
		return err
	}
	return nil
}

func (t *Task) clearDir() {
	if err := os.RemoveAll(t.scriptDir); err != nil {
		logx.Errorln(t.Name, t.workspace, t.scriptDir, err)
	}
}
