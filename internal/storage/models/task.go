package models

import (
	"time"
)

type Task struct {
	Name    string        `json:"name,omitempty" gorm:"not null;comment:名称"`
	Count   *int          `json:"count,omitempty" gorm:"comment:步骤数"`
	Timeout time.Duration `json:"timeout,omitempty" gorm:"not null;default:86400000000000;comment:超时时间"`
	TaskUpdate
}

type TaskUpdate struct {
	Message  string     `json:"message,omitempty" gorm:"comment:消息"`
	State    *int       `json:"state,omitempty" gorm:"not null;default:0;comment:状态"`
	OldState *int       `json:"old_state,omitempty" gorm:"not null;default:0;comment:旧状态"`
	STime    *time.Time `json:"s_time,omitempty" gorm:"comment:开始时间"`
	ETime    *time.Time `json:"e_time,omitempty" gorm:"comment:结束时间"`
}

type TaskEnv struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	Env
}
