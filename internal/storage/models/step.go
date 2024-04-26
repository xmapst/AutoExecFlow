package models

import (
	"time"
)

type Step struct {
	Name     string        `json:"name" gorm:"not null;comment:名称"`
	TaskName string        `json:"task_name" gorm:"not null;comment:任务名称"`
	Type     string        `json:"type" gorm:"not null;comment:类型"`
	Content  string        `json:"content" gorm:"comment:内容"`
	Timeout  time.Duration `json:"timeout" gorm:"not null;default:86400000000000;comment:超时时间"`
	StepUpdate
}

type StepUpdate struct {
	Message  string    `json:"message" gorm:"comment:消息"`
	State    *int      `json:"state" gorm:"not null;default:0;comment:状态"`
	OldState *int      `json:"old_state" gorm:"not null;default:0;comment:旧状态"`
	Code     *int64    `json:"code" gorm:"not null;default:0;comment:退出码"`
	STime    time.Time `json:"s_time" gorm:"comment:开始时间"`
	ETime    time.Time `json:"e_time" gorm:"comment:结束时间"`
}

type StepEnv struct {
	TaskName string `json:"task_name" gorm:"not null;comment:任务名称"`
	StepName string `json:"step_name" gorm:"not null;comment:步骤名称"`
	Env
}

type StepDepend struct {
	TaskName string `json:"task_name" gorm:"not null;comment:任务名称"`
	StepName string `json:"step_name" gorm:"not null;comment:步骤名称"`
	Name     string `json:"name" gorm:"not null;comment:依赖步骤名称"`
}
