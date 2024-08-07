package tables

import (
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type Task struct {
	Base
	models.Task
}

type TaskEnv struct {
	Base
	models.Env
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
}

type Step struct {
	Base
	models.Step
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
}

type StepEnv struct {
	Base
	models.Env
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
}

type StepDepend struct {
	Base
	models.StepDepend
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
}

type StepLog struct {
	Base
	models.Log
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
}
