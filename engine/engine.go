package engine

import (
	"errors"
	"fmt"
	"github.com/Jeffail/tunny"
	"github.com/natessilva/dag"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/config"
	"time"
)

type ExecTask struct {
	TaskID string
	Tasks  []*cache.Task
	State  *cache.TaskState
}

var (
	Pool    *tunny.Pool
	execErr = errors.New("exec error")
)

func NewExecPool(size int) {
	Pool = tunny.NewFunc(size, worker)
}

func Process(taskID, hardWareID, vmInstanceID string, tasks []*cache.Task) {
	// 临时存储
	var state = &cache.TaskState{
		State:        cache.Pending,
		Count:        int64(len(tasks)),
		HardWareID:   hardWareID,
		VMInstanceID: vmInstanceID,
		Times: &cache.Times{
			Begin: time.Now().UnixNano(),
			TTL:   config.App.KeyExpire,
		},
	}
	// 插入数据
	cache.SetTask(taskID, state, config.App.KeyExpire)
	Pool.Process(&ExecTask{
		TaskID: taskID,
		Tasks:  tasks,
		State:  state,
	})
}

func worker(i interface{}) interface{} {
	var e, ok = i.(*ExecTask)
	if !ok {
		logrus.Error("input problem")
		return nil
	}
	e.State.State = cache.Running
	cache.SetTask(e.TaskID, e.State, e.State.Times.TTL)

	// 编排步骤
	var runner = new(dag.Runner)

	// 创建一个有向无环图，图中的每个顶点都是一个作业
	for step, task := range e.Tasks {
		s := int64(step)
		t := task
		// 设置缓存中初始状态
		e.initStepCache(s, t)
		fn := func() error {
			return e.execStep(s, t)
		}
		runner.AddVertex(task.Name, fn)
	}
	// 根据在 depends_on 属性中配置的值创建顶点边
	for _, task := range e.Tasks {
		for _, dep := range task.DependsOn {
			runner.AddEdge(dep, task.Name)
		}
	}

	defer func() {
		e.State.Times.End = time.Now().UnixNano()
		// 更新数据
		cache.SetTask(e.TaskID, e.State, e.State.Times.TTL)
	}()

	if err := runner.Run(); err != nil {
		if err != execErr {
			logrus.Errorln(err)
			e.State.State = cache.SystemError
			e.State.Message = err.Error()
			return nil
		}
	}

	// 运行结束
	e.State.State = cache.Stop
	return nil
}

func (e *ExecTask) initStepCache(step int64, task *cache.Task) {
	var key = fmt.Sprintf("%s:%d_%s", e.TaskID, step, task.Name)
	var state = &cache.TaskStepState{
		Step:      step,
		Name:      task.Name,
		State:     cache.Pending,
		DependsOn: task.DependsOn,
		Message:   "如上一依赖步骤执行失败则一直保持待执行, 只有上一依赖步骤成功才会执行",
		Times: &cache.Times{
			TTL: e.State.Times.TTL,
		},
	}
	cache.SetTaskStep(key, state, state.Times.TTL)
}

func (e *ExecTask) newCmd(step int64, task *cache.Task) *Cmd {
	log := logrus.WithFields(logrus.Fields{
		"step":    step,
		"task_id": e.TaskID,
		"name":    task.Name,
		"shell":   task.CommandType,
		"envs":    task.EnvVars,
	})
	return &Cmd{
		Log:             log,
		TaskID:          e.TaskID,
		Step:            step,
		Name:            task.Name,
		Shell:           task.CommandType,
		Content:         task.CommandContent,
		ExternalEnvVars: task.EnvVars,
		Timeout:         task.Timeout,
		TTL:             e.State.Times.TTL,
	}
}

func (e *ExecTask) execStep(step int64, task *cache.Task) error {
	var key = fmt.Sprintf("%s:%d_%s", e.TaskID, step, task.Name)
	var state = &cache.TaskStepState{
		Step:      step,
		Name:      task.Name,
		State:     cache.Running,
		DependsOn: task.DependsOn,
		Times: &cache.Times{
			Begin: time.Now().UnixNano(),
			TTL:   e.State.Times.TTL,
		},
	}
	cache.SetTaskStep(key, state, state.Times.TTL)
	var cmd = e.newCmd(step, task)
	defer func() {
		state.Times.End = time.Now().UnixNano()
		state.State = cache.Stop
		cache.SetTaskStep(key, state, state.Times.TTL)
	}()
	if err := cmd.Create(); err != nil {
		cmd.Log.Error(err)
		state.Message = err.Error()
		state.Code = 255
		return execErr
	}
	state.Code, state.Message = cmd.Run()
	if state.Code != 0 {
		cmd.Log.Errorf("exit code is not 0 but %d", state.Code)
		return execErr
	}
	return nil
}
