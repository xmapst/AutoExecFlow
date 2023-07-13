package worker

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/dag"
	"github.com/xmapst/osreapi/internal/engine/manager"
	"github.com/xmapst/osreapi/internal/exec"
	"github.com/xmapst/osreapi/internal/logx"
)

type Task struct {
	log       logx.Logger
	workspace string
	ID        string
	State     *cache.TaskState
	Steps     []*cache.TaskStep
}

func Submit(task cache.Task) {
	// 临时存储
	var state = &cache.TaskState{
		State:    cache.Pending,
		Count:    int64(len(task.Steps)),
		MetaData: task.MetaData,
		Times: &cache.Times{
			ST: time.Now().UnixNano(),
			RT: config.App.KeyExpire,
		},
	}
	// 插入数据
	cache.SetTask(task.ID, state, config.App.KeyExpire)
	queue.PushBack(func() {
		workspace := filepath.Join(config.App.WorkSpace, task.ID)
		t := &Task{
			log: logx.GetSubLoggerWithKeyValue(map[string]string{
				"task":      task.ID,
				"workspace": workspace,
			}),
			workspace: workspace,
			ID:        task.ID,
			Steps:     task.Steps,
			State:     state,
		}
		ctx := manager.AddTask(context.Background(), task.ID)
		defer func(task string) {
			manager.LeaveTask(task)
		}(task.ID)
		res := t.run(ctx)
		t.log.Infoln("end of execution")
		if res != nil {
			t.log.Infoln(task.ID, res)
		}
	})
}

func (t *Task) run(ctx context.Context) error {
	defer func() {
		err := recover()
		if err != nil {
			t.log.Errorln(err)
		}
	}()

	t.State.State = cache.Running
	cache.SetTask(t.ID, t.State, t.State.Times.RT)

	defer func() {
		// 清理工作目录
		t.clear()
		// 结束时间
		t.State.Times.ET = time.Now().UnixNano()
		// 更新数据
		cache.SetTask(t.ID, t.State, t.State.Times.RT)
	}()

	if err := t.init(); err != nil {
		t.log.Errorln(err)
		t.State.State = cache.SystemError
		t.State.Message = err.Error()
		return nil
	}

	var stepFnMap = make(map[string]*dag.Step)
	for k, v := range t.Steps {
		sID := int64(k)
		step := v
		// 设置缓存中初始状态
		t.initStepCache(sID, step)
		fn := func(ctx context.Context) error {
			ctx = manager.AddTaskStep(ctx, t.ID, sID)
			defer func(task string, step int64) {
				manager.LeaveTaskStep(task, step)
			}(t.ID, sID)
			return t.execStep(ctx, sID, step)
		}
		stepFnMap[step.Name] = dag.NewStep(step.Name, fn)
	}

	// 编排步骤: 创建一个有向无环图，图中的每个顶点都是一个作业
	var flow = dag.NewTask()
	for _, step := range t.Steps {
		stepFn, ok := stepFnMap[step.Name]
		if !ok {
			continue
		}
		// 添加顶点以及设置依赖关系
		flow.Add(stepFn).WithDeps(func() []*dag.Step {
			var stepFns []*dag.Step
			for _, name := range step.DependsOn {
				_stepFn, _ok := stepFnMap[name]
				if !_ok {
					continue
				}
				stepFns = append(stepFns, _stepFn)
			}
			return stepFns
		}()...)
	}

	var state = cache.Stop

	defer func() {
		t.State.State = state
	}()
	if err := flow.Run(ctx); err != nil {
		if err == dag.ErrCycleDetected || err == dag.ErrEmptyTask {
			state = cache.SystemError
		}
		t.State.Message = err.Error()
		t.log.Errorln(err)
	}
	return nil
}

func (t *Task) init() error {
	err := os.MkdirAll(t.workspace, 0777)
	if err != nil && err != os.ErrExist {
		t.log.Errorln(err)
		return err
	}
	return nil
}

func (t *Task) clear() {
	t.log.Infof("cleanup workspace")
	err := os.RemoveAll(t.workspace)
	if err != nil {
		t.log.Errorln(err)
	}
}

func (t *Task) initStepCache(step int64, task *cache.TaskStep) {
	var state = &cache.TaskStepState{
		Step:      step,
		Name:      task.Name,
		State:     cache.Pending,
		DependsOn: task.DependsOn,
		Message:   "The current step only proceeds if the previous step succeeds.",
		Times: &cache.Times{
			RT: t.State.Times.RT,
		},
	}
	cache.SetTaskStep(t.ID, step, state, state.Times.RT)
}

func (t *Task) newCmd(id int64, step *cache.TaskStep) *exec.Cmd {
	return &exec.Cmd{
		TaskID:          t.ID,
		StepID:          id,
		Name:            step.Name,
		Shell:           step.CommandType,
		Content:         step.CommandContent,
		Workspace:       t.workspace,
		ScriptDir:       config.App.ScriptDir,
		ExternalEnvVars: step.EnvVars,
		Timeout:         step.Timeout,
		TTL:             t.State.Times.RT,
	}
}

func (t *Task) execStep(ctx context.Context, step int64, task *cache.TaskStep) (err error) {
	var state = &cache.TaskStepState{
		Step:      step,
		Name:      task.Name,
		State:     cache.Running,
		DependsOn: task.DependsOn,
		Times: &cache.Times{
			ST: time.Now().UnixNano(),
			RT: t.State.Times.RT,
		},
	}
	cache.SetTaskStep(t.ID, step, state, state.Times.RT)
	var cmd = t.newCmd(step, task)
	defer func() {
		state.Times.ET = time.Now().UnixNano()
		state.State = cache.Stop
		cache.SetTaskStep(t.ID, step, state, state.Times.RT)
	}()
	if err = cmd.Create(); err != nil {
		t.log.Errorln(err)
		state.Message = err.Error()
		state.Code = 255
		return
	}
	state.Code, err = cmd.Run(ctx)
	if err != nil {
		t.log.Errorln(err)
		state.Message = err.Error()
		return err
	}
	state.Message = "execution succeed"
	return
}
