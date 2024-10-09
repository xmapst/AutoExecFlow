package models

type TaskEnv struct {
	Base
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	Env
}
