package models

type StepEnv struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
	Env
}
