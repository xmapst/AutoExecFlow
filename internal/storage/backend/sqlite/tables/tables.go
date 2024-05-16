package tables

import (
	"github.com/xmapst/osreapi/internal/storage/models"
)

type Task struct {
	models.Task
	Base
}

type TaskEnv struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	models.Env
	Base
}

type Step struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	models.Step
	Base
}

type StepEnv struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"not null;comment:步骤名称"`
	models.Env
	Base
}

type StepDepend struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"not null;comment:步骤名称"`
	Name     string `json:"name" gorm:"not null;comment:依赖步骤名称"`
	Base
}

type Log struct {
	TaskName string `json:"task_name,omitempty" gorm:"not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"not null;comment:步骤名称"`
	models.Log
	Base
}
