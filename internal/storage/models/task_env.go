package models

type STaskEnv struct {
	SBase
	TaskName string `json:"task_name,omitempty" gorm:"size:256;index;not null;comment:任务名称"`
	SEnv
}

func (s *STaskEnv) TableName() string {
	return "t_task_env"
}
