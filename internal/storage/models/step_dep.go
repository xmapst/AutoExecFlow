package models

type SStepDepend struct {
	SBase
	TaskName string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	StepName string `json:"step_name,omitempty" gorm:"index;not null;comment:步骤名称"`
	Name     string `json:"name,omitempty" gorm:"not null,comment:依赖步骤名称"`
}

func (s *SStepDepend) TableName() string {
	return "t_step_depend"
}
