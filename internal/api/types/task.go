package types

import (
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

const (
	XTaskName  = "X-Task-Name"
	XTaskState = "X-Task-STATE"
)

// 只允许中文,英文(含大小写),0-9,-_.~字符
var reg = regexp.MustCompile("[^a-zA-Z\\p{Han}0-9\\-_.~]")

// Response struct

type TaskCreateRes struct {
	// 链接
	URL string `json:"url" yaml:"URL"`
	// 任务名称
	Name string `json:"name" yaml:"Name"`
	// 步骤数量
	Count int `json:"count" yaml:"Count"`
	// Deprecated, use Name
	ID string `json:"id" yaml:"ID" swaggerignore:"true"`
}

type TaskListRes struct {
	// 分页
	Page Page `json:"page" yaml:"Page"`
	// 任务列表
	Tasks []*TaskRes `json:"tasks" yaml:"Tasks"`
}

type Page struct {
	// 当前页
	Current int64 `json:"current" yaml:"Current"`
	// 页大小
	Size int64 `json:"size" yaml:"Size"`
	// 总页数
	Total int64 `json:"total" yaml:"Total"`
}

type TaskRes struct {
	// 任务名称
	Name string `json:"name" yaml:"Name"`
	// 任务状态
	State string `json:"state" yaml:"State"`
	// 任务管理
	Manager string `json:"manager" yaml:"Manager"`
	// 任务工作空间
	Workspace string `json:"workspace" yaml:"Workspace"`
	// 任务信息
	Message string `json:"msg" yaml:"Message"`
	// 步骤数量
	Count int `json:"count" yaml:"Count"`
	// 任务环境变量
	Env map[string]string `json:"env,omitempty" yaml:"Env,omitempty"`
	// 任务超时时间
	Timeout string `json:"timeout,omitempty" yaml:"Timeout,omitempty"`
	// 是否禁用
	Disable bool `json:"disable" yaml:"Disable"`
	// 时间
	Time *Time `json:"time,omitempty" yaml:"Time,omitempty"`
}

type TaskStepRes struct {
	// 步骤名称
	Name string `json:"name" yaml:"Name"`
	// 步骤退出码
	Code int64 `json:"code" yaml:"Code"`
	// 步骤状态
	State string `json:"state" yaml:"State"`
	// 步骤管理
	Manager string `json:"manager" yaml:"Manager"`
	// 步骤工作空间
	Workspace string `json:"workspace" yaml:"Workspace"`
	// 步骤信息
	Message string `json:"msg" yaml:"Message"`
	// 步骤超时时间
	Timeout string `json:"timeout,omitempty" yaml:"Timeout,omitempty"`
	// 是否禁用
	Disable bool `json:"disable" yaml:"Disable"`
	// 步骤依赖
	Depends []string `json:"depends,omitempty" yaml:"Depends,omitempty"`
	// 步骤环境变量
	Env map[string]string `json:"env,omitempty" yaml:"Env,omitempty"`
	// 步骤类型
	Type string `json:"type,omitempty" yaml:"Type,omitempty"`
	// 步骤内容
	Content string `json:"content,omitempty" yaml:"Content,omitempty"`
	// 时间
	Time *Time `json:"time,omitempty" yaml:"Time,omitempty"`
}

type Time struct {
	// 开始时间
	Start string `json:"start,omitempty" yaml:"Start,omitempty"`
	// 结束时间
	End string `json:"end,omitempty" yaml:"End,omitempty"`
}

type TaskStepLogRes struct {
	// 时间戳
	Timestamp int64 `json:"timestamp" yaml:"Timestamp"`
	// 行号
	Line int64 `json:"line" yaml:"Line"`
	// 内容
	Content string `json:"content" yaml:"Content"`
}

// Request struct

type TaskStep struct {
	// 步骤名称
	Name string `json:"name" form:"name" yaml:"Name" example:"script.ps1"`
	// 步骤超时时间
	Type string `json:"type" form:"type" yaml:"Type" example:"powershell"`
	// 步骤内容
	Content string `json:"content" form:"content" yaml:"Content" example:"sleep 10"`
	// 步骤环境变量
	Env map[string]string `json:"env" form:"env" yaml:"Env" example:"key:value,key1:value1"`
	// 步骤依赖
	Depends []string `json:"depends" form:"depends" yaml:"Depends" example:""`
	// 步骤超时时间
	Timeout string `json:"timeout" form:"timeout" yaml:"Timeout" example:"3m"`
	// 是否禁用
	Disable bool `json:"disable" form:"disable" yaml:"Disable" example:"false"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars" form:"env_vars" yaml:"EnvVars" example:"env1=value1,env2=value2" swaggerignore:"true"`
	// Deprecated, use Type
	CommandType string `json:"command_type" form:"command_type" yaml:"CommandType" example:"powershell" swaggerignore:"true"`
	// Deprecated, use Content
	CommandContent string `json:"command_content" form:"command_content" yaml:"CommandContent" example:"sleep 10" swaggerignore:"true"`

	timeoutDuration time.Duration
}

type TaskSteps []TaskStep

func (s TaskSteps) review(timeout time.Duration, async bool) error {
	for k, v := range s {
		if v.CommandType != "" && v.Type == "" {
			v.Type = v.CommandType
			s[k].Type = v.Type
		}
		if v.Type == "" {
			return errors.New("key: 'Step.Type' Error:Field validation for 'Type' failed on the 'required' tag")
		}

		if v.CommandContent != "" && v.Content == "" {
			v.Content = v.CommandContent
			s[k].Content = v.Content
		}
		if v.Content == "" {
			return errors.New("key: 'Step.Content' Error:Field validation for 'Content' failed on the 'required' tag")
		}
	}

	// 处理旧环境变量接收方式
	for k, v := range s {
		if s[k].Env == nil {
			s[k].Env = make(map[string]string)
		}
		for kk, vv := range utils.SliceToStrMap(v.EnvVars) {
			if _, ok := v.Env[kk]; !ok {
				s[k].Env[kk] = vv
			}
		}
	}

	s.parseDuration(timeout)
	s.fixName()
	if err := s.uniqNames(); err != nil {
		logx.Errorln(err)
		return err
	}
	if !async {
		// 非编排模式,按顺序执行
		s.fixSync()
	}
	return nil
}

func (s TaskSteps) fixName() {
	name := ksuid.New().String()
	for k, v := range s {
		v.Name = reg.ReplaceAllString(v.Name, "")
		if v.Name == "" {
			v.Name = fmt.Sprintf("%s-%d", name, k+1)
		}
		s[k].Name = v.Name
	}
}

func (s TaskSteps) fixSync() {
	for k := range s {
		if k == 0 {
			s[k].Depends = nil
			continue
		}
		s[k].Depends = []string{s[k-1].Name}
	}
}

func (s TaskSteps) uniqNames() error {
	counts := make(map[string]int)
	for _, v := range s {
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

func (s TaskSteps) parseDuration(t time.Duration) {
	for k, v := range s {
		timeout, _ := time.ParseDuration(v.Timeout)
		if timeout > t || timeout <= 0 {
			timeout = t
		}
		s[k].timeoutDuration = timeout
	}
}

type Task struct {
	// 任务名称
	Name string `json:"name" query:"name"  form:"name" yaml:"Name" example:"task_name"`
	// 任务超时时间
	Timeout string `json:"timeout" query:"timeout" form:"timeout" yaml:"Timeout" example:"24h"`
	// 任务环境变量
	Env map[string]string `json:"env" query:"env" form:"env" yaml:"Env" example:"key:value,key1:value1"`
	// 是否异步执行
	Async bool `json:"async" query:"async" form:"async" yaml:"Async" example:"false"`
	// 是否禁用
	Disable bool `json:"disable" query:"disable" form:"disable" yaml:"Disable" example:"false"`
	// 任务步骤
	Step TaskSteps `json:"step" form:"step" yaml:"Step"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars" query:"env_vars" form:"env_vars" yaml:"EnvVars" example:"" swaggerignore:"true"`

	timeoutDuration time.Duration
}

func (t *Task) review() error {
	if t.Step == nil || len(t.Step) == 0 {
		return errors.New("key: 'Task.Step' Error:Field validation for 'Step' failed on the 'required' tag")
	}
	if t.Env == nil {
		t.Env = make(map[string]string)
	}
	// 处理旧环境变量接收方式
	for k, v := range utils.SliceToStrMap(t.EnvVars) {
		if _, ok := t.Env[k]; !ok {
			t.Env[k] = v
		}
	}
	t.Name = reg.ReplaceAllString(t.Name, "")
	if t.Name == "" {
		t.Name = ksuid.New().String()
	}

	timeout, err := time.ParseDuration(t.Timeout)
	if err != nil {
		timeout = config.App.ExecTimeOut
	}
	t.timeoutDuration = timeout
	if err = t.Step.review(t.timeoutDuration, t.Async); err != nil {
		return err
	}
	return nil
}

func (t *Task) Save() (err error) {
	// 检查请求内容
	if err = t.review(); err != nil {
		return err
	}

	// 检查任务是否在运行
	if _, err = dag.GraphManager(t.Name); err == nil {
		return errors.New("task is running")
	}

	var task = storage.Task(t.Name)
	// 清理旧数据
	task.ClearAll()

	defer func() {
		if err != nil {
			// rollback
			task.ClearAll()
		}
	}()

	// save task
	err = task.Insert(&models.Task{
		Count:   models.Pointer(len(t.Step)),
		Timeout: t.timeoutDuration,
		Disable: models.Pointer(t.Disable),
		TaskUpdate: models.TaskUpdate{
			Message:  "the task is waiting to be scheduled for execution",
			State:    models.Pointer(models.Pending),
			OldState: models.Pointer(models.Pending),
		},
	})
	if err != nil {
		return err
	}

	// save task env
	for name, value := range t.Env {
		if err = task.Env().Insert(&models.Env{
			Name:  name,
			Value: value,
		}); err != nil {
			return err
		}
	}

	for _, step := range t.Step {
		// save step
		err = task.Step(step.Name).Insert(&models.Step{
			Type:    step.Type,
			Content: step.Content,
			Timeout: step.timeoutDuration,
			Disable: models.Pointer(step.Disable),
			StepUpdate: models.StepUpdate{
				Message:  "the step is waiting to be scheduled for execution",
				Code:     models.Pointer(int64(0)),
				State:    models.Pointer(models.Pending),
				OldState: models.Pointer(models.Pending),
			},
		})
		if err != nil {
			return err
		}
		// save step env
		for name, value := range step.Env {
			if err = task.Step(step.Name).Env().Insert(&models.Env{
				Name:  name,
				Value: value,
			}); err != nil {
				return err
			}
		}
		// save step depend
		err = task.Step(step.Name).Depend().Insert(step.Depends...)
		if err != nil {
			return err
		}
	}
	// 提交任务
	return worker.Submit(t.Name)
}
