package models

import (
	"time"
)

type Task struct {
	Name    string        `json:"name,omitempty" gorm:"index;not null;comment:名称"`
	Count   *int          `json:"count,omitempty" gorm:"default:0;comment:步骤数"`
	Timeout time.Duration `json:"timeout,omitempty" gorm:"not null;default:86400000000000;comment:超时时间"`
	Disable *bool         `json:"disable,omitempty" gorm:"not null;default:false;comment:禁用"`
	TaskUpdate
}

type TaskUpdate struct {
	Message  string     `json:"message,omitempty" gorm:"comment:消息"`
	State    *int       `json:"state,omitempty" gorm:"index;not null;default:0;comment:状态"`
	OldState *int       `json:"old_state,omitempty" gorm:"index;not null;default:0;comment:旧状态"`
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
