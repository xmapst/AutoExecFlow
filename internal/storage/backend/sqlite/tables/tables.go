package tables

import (
	"github.com/xmapst/osreapi/internal/storage/models"
)

type Task struct {
	Base
	models.Task
}

type TaskEnv struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	models.Env
}

type Step struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	models.Step
}

type StepEnv struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
	models.Env
}

type StepDepend struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
	Name     string `json:"name" gorm:"not null,comment:依赖步骤名称"`
}

type StepLog struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
	models.Log
}
