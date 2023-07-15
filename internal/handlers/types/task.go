package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/engine/manager"
	"github.com/xmapst/osreapi/internal/logx"
	"github.com/xmapst/osreapi/internal/utils"
)

const (
	XTaskID    = "X-Task-ID"
	XTaskState = "X-Task-STATE"
)

type TaskDetailRes struct {
	ID        int64    `json:"id"`
	URL       string   `json:"url,omitempty"`
	Name      string   `json:"name,omitempty"`
	State     int      `json:"state"`
	Code      int64    `json:"code"`
	Message   string   `json:"msg"`
	DependsOn []string `json:"depends_on,omitempty"`
	Times     *Times   `json:"times,omitempty"`
}

type TaskListRes struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	State   int    `json:"state"`
	Code    int64  `json:"code"`
	Count   int64  `json:"count"`
	Message string `json:"msg"`
	Times   *Times `json:"times"`
}

type Times struct {
	ST string `json:"st,omitempty"` // 开始时间
	ET string `json:"et,omitempty"` // 结束时间
	RT string `json:"rt,omitempty"` // 剩余时间
}

type Step struct {
	CommandType    string   `json:"command_type" form:"command_type" binding:"required" example:"powershell"`
	CommandContent string   `json:"command_content" form:"command_content" binding:"required" example:"sleep 10"`
	Name           string   `json:"name,omitempty" form:"name,omitempty" example:"script.ps1"`
	EnvVars        []string `json:"env_vars,omitempty" form:"env_vars,omitempty" example:"env1=value1,env2=value2"`
	DependsOn      []string `json:"depends_on,omitempty" form:"depends_on,omitempty"  example:""`
	Timeout        string   `json:"timeout,omitempty" form:"timeout,omitempty" example:"3m"`
	Notify         *Notifys `json:"notify" form:"notify"`

	TimeoutDuration time.Duration `json:"-" form:"-"`
}

type Notifys []*Notify

type Notify struct {
	Type   string `json:"type" form:"type" binding:"required" example:"webhook"`
	Action string `json:"action" form:"action" binding:"required" example:"before"` // or after
}

type Steps []*Step

func (s Steps) Check(id string, ansync bool) error {
	for _, v := range s {
		if v.CommandType == "" {
			return errors.New("key: 'Step.CommandType' Error:Field validation for 'CommandType' failed on the 'required' tag")
		}
		if v.CommandContent == "" {
			return errors.New("key: 'Step.CommandContent' Error:Field validation for 'CommandContent' failed on the 'required' tag")
		}
	}
	s.parseDuration()
	s.fixName(id)
	if err := s.uniqNames(); err != nil {
		logx.Errorln(err)
		return err
	}
	if !ansync {
		// 非编排模式,按顺序执行
		s.fixSync()
	}
	return nil
}

func (s Steps) fixName(taskID string) {
	for step, v := range s {
		if v.Name == "" {
			s[step].Name = fmt.Sprintf("%s-%d", taskID, step)
		}
	}
}

func (s Steps) fixSync() {
	for k := range s {
		if k == 0 {
			s[k].DependsOn = nil
			continue
		}
		s[k].DependsOn = []string{s[k-1].Name}
	}
}

func (s Steps) uniqNames() (result error) {
	counts := make(map[string]int)
	for _, v := range s {
		counts[v.Name]++
	}
	for name, count := range counts {
		if count > 1 {
			result = multierror.Append(result, fmt.Errorf("%s repeat count %d", name, count))
		}
	}
	return
}

func (s Steps) parseDuration() {
	for k, v := range s {
		timeout, err := time.ParseDuration(v.Timeout)
		if v.Timeout == "" || err != nil {
			timeout = config.App.ExecTimeOut
		}
		s[k].TimeoutDuration = timeout
	}
}

func (s Steps) GetMetaData() (res cache.MetaData) {
	// check envs
	for k, v := range s {
		var _env []string
		m := utils.SliceToStrMap(v.EnvVars)
		for k, v := range m {
			_env = append(_env, fmt.Sprintf("%s=%s", k, v))
		}
		str, ok := m["HARDWARE_ID"]
		if ok && str != "" {
			res.HardWareID = str
		}
		str, ok = m["VM_INSTANCE_ID"]
		if ok && str != "" {
			res.VMInstanceID = str
		}
		s[k].EnvVars = _env
	}
	return
}

type Task struct {
	ID      string  `query:"id" json:"id" form:"id" example:""`
	Timeout string  `query:"timeout" json:"timeout" form:"timeout" example:""`
	AnSync  bool    `query:"ansync" json:"ansync" form:"ansync" example:"false"`
	Step    Steps   `json:"step" form:"step"`
	Notify  Notifys `json:"notify" form:"notify"`

	TimeoutDuration time.Duration `json:"-" form:"-"`
}

func (t *Task) Check() error {
	if t.Step == nil || len(t.Step) == 0 {
		return errors.New("key: 'Task.Step' Error:Field validation for 'Step' failed on the 'required' tag")
	}
	if manager.TaskRunning(t.ID) {
		return errors.New("task is running")
	}
	if t.ID == "" {
		t.ID = ksuid.New().String()
	}
	timeout, err := time.ParseDuration(t.Timeout)
	if err == nil {
		t.TimeoutDuration = timeout
	}
	if err := t.Step.Check(t.ID, t.AnSync); err != nil {
		return err
	}
	return nil
}
