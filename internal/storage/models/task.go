package models

import (
	"time"
)

type Task struct {
	Base
	Name        string        `json:"name,omitempty" gorm:"index;not null;comment:名称"`
	Description string        `json:"description,omitempty" gorm:"comment:描述"`
	Node        string        `json:"node,omitempty" gorm:"index;default:null;comment:节点"`
	Async       *bool         `json:"async,omitempty" gorm:"not null;default:false;comment:异步"`
	Timeout     time.Duration `json:"timeout,omitempty" gorm:"not null;default:86400000000000;comment:超时时间"`
	Disable     *bool         `json:"disable,omitempty" gorm:"not null;default:false;comment:禁用"`
	TaskUpdate
}

type TaskUpdate struct {
	Message  string     `json:"message,omitempty" gorm:"comment:消息"`
	State    *State     `json:"state,omitempty" gorm:"index;not null;default:0;comment:状态"`
	OldState *State     `json:"old_state,omitempty" gorm:"index;not null;default:0;comment:旧状态"`
	STime    *time.Time `json:"s_time,omitempty" gorm:"comment:开始时间"`
	ETime    *time.Time `json:"e_time,omitempty" gorm:"comment:结束时间"`
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
