package models

type SStepEnv struct {
	SBase
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
	SEnv
}

func (s *SStepEnv) TableName() string {
	return "t_step_env"
}
