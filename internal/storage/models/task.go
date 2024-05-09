package models

import (
	"time"
)

type Task struct {
	Name    string        `json:"name,omitempty"`
	Count   *int          `json:"count,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
	TaskUpdate
}

type TaskUpdate struct {
	Message  string     `json:"message,omitempty"`
	State    *int       `json:"state,omitempty"`
	OldState *int       `json:"old_state,omitempty"`
	STime    *time.Time `json:"s_time,omitempty"`
	ETime    *time.Time `json:"e_time,omitempty"`
}

func (t *TaskUpdate) STimeStr() string {
	if t.STime == nil {
		return ""
	}
	return t.STime.Format(time.RFC3339)
}

func (t *TaskUpdate) ETimeStr() string {
	if t.ETime == nil {
		return ""
	}
	return t.ETime.Format(time.RFC3339)
}

type Tasks []*Task

func (t Tasks) Len() int {
	return len(t)
}

func (t Tasks) Less(i, j int) bool {
	if t[i].STime == nil || t[j].STime == nil {
		return t[i].Name < t[j].Name
	}
	iTime := t[i].STime.Nanosecond()
	jTime := t[j].STime.Nanosecond()
	return iTime < jTime
}

func (t Tasks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type TaskEnv struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	Env
}
