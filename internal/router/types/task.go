package types

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/logx"
)

const (
	XTaskName  = "X-Task-Name"
	XTaskState = "X-Task-STATE"
)

var reg = regexp.MustCompile("[^a-zA-Z\\p{Han}0-9]")

// Response struct

type TaskCreateRes struct {
	URL   string `json:"url"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	// Deprecated, use Name
	ID string `json:"id"`
}

type TaskListRes struct {
	Page  Page       `json:"page"`
	Tasks []*TaskRes `json:"tasks"`
}

type Page struct {
	// 当前页
	Current int64 `json:"current"`
	// 页大小
	Size int64 `json:"size"`
	// 总页数
	Total int64 `json:"total"`
}

type TaskRes struct {
	Name      string            `json:"name"`
	State     string            `json:"state"`
	Manager   string            `json:"manager"`
	Workspace string            `json:"workspace"`
	Message   string            `json:"msg"`
	Count     int               `json:"count"`
	Env       map[string]string `json:"env"`
	Timeout   string            `json:"timeout"`
	Time      *Time             `json:"time"`
}

type StepRes struct {
	Name      string            `json:"name"`
	Code      int64             `json:"code"`
	State     string            `json:"state"`
	Manager   string            `json:"manager"`
	Workspace string            `json:"workspace"`
	Message   string            `json:"msg"`
	Timeout   string            `json:"timeout"`
	Depends   []string          `json:"depends"`
	Env       map[string]string `json:"env"`
	Type      string            `json:"type"`
	Content   string            `json:"content"`
	Time      *Time             `json:"time"`
}

type Time struct {
	Start string `json:"start"` // 开始时间
	End   string `json:"end"`   // 结束时间
}

type LogRes struct {
	Timestamp int64  `json:"timestamp"`
	Line      int64  `json:"line"`
	Content   string `json:"content"`
}

// Request struct

type Step struct {
	Name    string            `json:"name" example:"script.ps1"`
	Type    string            `json:"type" example:"powershell"`
	Content string            `json:"content" example:"sleep 10"`
	Env     map[string]string `json:"env" example:"key:value;key1:value1"`
	Depends []string          `json:"depends" example:""`
	Timeout string            `json:"timeout" example:"3m"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars" example:"env1=value1,env2=value2"`
	// Deprecated, use Type
	CommandType string `json:"command_type" example:"powershell"`
	// Deprecated, use Content
	CommandContent string `json:"command_content" example:"sleep 10"`

	timeoutDuration time.Duration
}

type Steps []*Step

func (s Steps) review(timeout time.Duration, async bool) error {
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

func (s Steps) fixName() {
	name := ksuid.New().String()
	for k, v := range s {
		v.Name = reg.ReplaceAllString(v.Name, "")
		if v.Name == "" {
			v.Name = fmt.Sprintf("%s-%d", name, k+1)
		}
		s[k].Name = v.Name
	}
}

func (s Steps) fixSync() {
	for k := range s {
		if k == 0 {
			s[k].Depends = nil
			continue
		}
		s[k].Depends = []string{s[k-1].Name}
	}
}

func (s Steps) uniqNames() error {
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

func (s Steps) parseDuration(t time.Duration) {
	for k, v := range s {
		timeout, _ := time.ParseDuration(v.Timeout)
		if timeout > t || timeout <= 0 {
			timeout = t
		}
		s[k].timeoutDuration = timeout
	}
}

type Task struct {
	Name    string            `json:"name" example:""`
	Timeout string            `json:"timeout" example:""`
	Env     map[string]string `json:"env" example:"key:value;key1:value1"`
	Async   bool              `json:"async" example:"false"`
	Step    Steps             `json:"step"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars" example:""`

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
	worker.Submit(t.Name)
	return
}
