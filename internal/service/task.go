package service

import (
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// 只允许中文,英文(含大小写),0-9,-_.~字符
var reg = regexp.MustCompile("[^a-zA-Z\\p{Han}0-9\\-_.~]")

type TaskService struct {
	name string
}

func Task(name string) *TaskService {
	return &TaskService{
		name: name,
	}
}

func TaskList(req *types.PageReq) *types.TaskListRes {
	tasks, total := storage.TaskList(req.Page, req.Size, req.Prefix)
	if tasks == nil {
		return nil
	}
	pageTotal := total / req.Size
	if total%req.Size != 0 {
		pageTotal += 1
	}
	var list = &types.TaskListRes{
		Page: types.PageRes{
			Current: req.Page,
			Size:    req.Size,
			Total:   pageTotal,
		},
	}
	for _, task := range tasks {
		res := &types.TaskRes{
			Name:    task.Name,
			State:   models.StateMap[*task.State],
			Message: task.Message,
			Env:     make(map[string]string),
			Timeout: task.Timeout.String(),
			Disable: *task.Disable,
			Count:   *task.Count,
			Time: &types.TimeRes{
				Start: task.STimeStr(),
				End:   task.ETimeStr(),
			},
		}
		st := storage.Task(task.Name)
		// 获取任务级所有环境变量
		envs := st.Env().List()
		for _, env := range envs {
			res.Env[env.Name] = env.Value
		}

		// 获取当前进行到那些步骤
		steps := st.StepList(backend.All)
		var groups = make(map[models.State][]string)
		for _, v := range steps {
			groups[*v.State] = append(groups[*v.State], v.Name)
		}
		res.Message = GenerateStateMessage(res.Message, groups)
		list.Tasks = append(list.Tasks, res)
	}
	return list
}

func (ts *TaskService) Create(task *types.TaskReq) (err error) {
	// 检查任务是否在运行
	if _, err = dag.GraphManager(task.Name); err == nil {
		return errors.New("task is running")
	}

	// 检查请求内容
	timeout, err := ts.review(task)
	if err != nil {
		logx.Errorln(err)
		return err
	}

	var db = storage.Task(task.Name)
	// 清理旧数据
	db.ClearAll()

	defer func() {
		if err != nil {
			// rollback
			db.ClearAll()
		}
	}()

	if timeout, err = ts.saveTask(db, timeout, task); err != nil {
		return err
	}

	err = ts.reviewStep(task.Async, task.Step)
	if err != nil {
		return err
	}

	for _, step := range task.Step {
		// save step
		stepSvc := Step(task.Name, step.Name)
		if err = stepSvc.Create(timeout, step); err != nil {
			return err
		}
	}
	// 提交任务
	return worker.Submit(task.Name)
}

func (ts *TaskService) review(task *types.TaskReq) (time.Duration, error) {
	if task.Step == nil || len(task.Step) == 0 {
		return 0, errors.New("key: 'Task.Step' Error:Field validation for 'Step' failed on the 'required' tag")
	}
	if task.Env == nil {
		task.Env = make(map[string]string)
	}
	// 处理旧环境变量接收方式
	for k, v := range utils.SliceToStrMap(task.EnvVars) {
		if _, ok := task.Env[k]; !ok {
			task.Env[k] = v
		}
	}
	task.Name = reg.ReplaceAllString(task.Name, "")
	if task.Name == "" {
		task.Name = ksuid.New().String()
	}
	timeout, err := time.ParseDuration(task.Timeout)
	if err != nil {
		timeout = config.App.ExecTimeOut
	}
	return timeout, nil
}

func (ts *TaskService) reviewStep(async bool, steps types.TaskStepsReq) error {
	// 检查步骤名称是否重复
	if err := ts.uniqStepsName(steps); err != nil {
		return err
	}
	if !async {
		// 非编排模式,按顺序执行
		for k := range steps {
			if k == 0 {
				steps[k].Depends = nil
				continue
			}
			steps[k].Depends = []string{steps[k-1].Name}
		}
	}
	return nil
}

func (ts *TaskService) uniqStepsName(steps types.TaskStepsReq) error {
	counts := make(map[string]int)
	for _, v := range steps {
		counts[v.Name]++
	}
	var errs []error
	for name, count := range counts {
		if count > 1 {
			errs = append(errs, fmt.Errorf("%s repeat count %d", name, count))
		}
	}
	if errs == nil {
		return nil
	}
	return fmt.Errorf("%v", errs)
}
func (ts *TaskService) saveTask(db backend.ITask, timeout time.Duration, task *types.TaskReq) (time.Duration, error) {
	// save task
	err := db.Insert(&models.Task{
		Count:   models.Pointer(len(task.Step)),
		Timeout: timeout,
		Disable: models.Pointer(task.Disable),
		TaskUpdate: models.TaskUpdate{
			Message:  "the task is waiting to be scheduled for execution",
			State:    models.Pointer(models.Pending),
			OldState: models.Pointer(models.Pending),
		},
	})
	if err != nil {
		return timeout, err
	}

	// save task env
	for name, value := range task.Env {
		if err = db.Env().Insert(&models.Env{
			Name:  name,
			Value: value,
		}); err != nil {
			return timeout, err
		}
	}
	return timeout, nil
}

func (ts *TaskService) Delete() error {
	storage.Task(ts.name).ClearAll()
	return nil
}

func (ts *TaskService) Detail() (types.Code, *types.TaskRes, error) {
	db := storage.Task(ts.name)
	task, err := db.Get()
	if err != nil {
		logx.Errorln(err)
		return types.CodeFailed, nil, err
	}

	data := &types.TaskRes{
		Name:    task.Name,
		State:   models.StateMap[*task.State],
		Message: task.Message,
		Env:     make(map[string]string),
		Timeout: task.Timeout.String(),
		Disable: *task.Disable,
		Count:   *task.Count,
		Time: &types.TimeRes{
			Start: task.STimeStr(),
			End:   task.ETimeStr(),
		},
	}
	for _, env := range db.Env().List() {
		data.Env[env.Name] = env.Value
	}

	// 获取当前进行到那些步骤
	steps := db.StepList(backend.All)
	var groups = make(map[models.State][]string)
	for _, v := range steps {
		groups[*v.State] = append(groups[*v.State], v.Name)
	}
	data.Message = GenerateStateMessage(data.Message, groups)
	return ConvertState(*task.State), data, nil
}

func (ts *TaskService) Manager(action string, duration string) error {
	manager, err := dag.GraphManager(ts.name)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	task, err := storage.Task(ts.name).Get()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	if *task.State != models.Running && *task.State != models.Pending && *task.State != models.Paused {
		return errors.New("task is no running")
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			return storage.Task(ts.name).Update(&models.TaskUpdate{
				State:    models.Pointer(models.Failed),
				OldState: task.State,
				Message:  "has been killed",
			})
		}
	case "pause":
		if manager.State() != dag.StatePaused {
			_ = manager.Pause(duration)
			return storage.Task(ts.name).Update(&models.TaskUpdate{
				State:    models.Pointer(models.Paused),
				OldState: task.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.State() == dag.StatePaused {
			manager.Resume()
			return storage.Task(ts.name).Update(&models.TaskUpdate{
				State:    task.OldState,
				OldState: task.State,
				Message:  "has been resumed",
			})
		}
	}
	return nil
}

func (ts *TaskService) Steps() (code types.Code, data []*types.TaskStepRes, err error) {
	db := storage.Task(ts.name)
	task, err := db.Get()
	if err != nil {
		return types.CodeNoData, nil, err
	}
	steps := db.StepList(backend.All)
	if steps == nil {
		return types.CodeNoData, nil, errors.New("steps not found")
	}

	var groups = make(map[models.State][]string)
	for _, step := range steps {
		groups[*step.State] = append(groups[*step.State], step.Name)
		res := &types.TaskStepRes{
			Name:    step.Name,
			State:   models.StateMap[*step.State],
			Code:    *step.Code,
			Message: step.Message,
			Timeout: step.Timeout.String(),
			Disable: *step.Disable,
			Env:     make(map[string]string),
			Type:    step.Type,
			Content: step.Content,
			Time: &types.TimeRes{
				Start: step.STimeStr(),
				End:   step.ETimeStr(),
			},
		}
		res.Depends = db.Step(step.Name).Depend().List()
		envs := db.Step(step.Name).Env().List()
		for _, env := range envs {
			res.Env[env.Name] = env.Value
		}
		data = append(data, res)
	}
	task.Message = GenerateStateMessage(task.Message, groups)
	return ConvertState(*task.State), data, errors.New(task.Message)
}
