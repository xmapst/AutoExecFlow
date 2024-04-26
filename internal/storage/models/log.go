package models

type Log struct {
	TaskName  string `json:"task_name" gorm:"not null;comment:任务名称"`
	StepName  string `json:"step_name" gorm:"not null;comment:步骤名称"`
	Timestamp int64  `json:"timestamp" gorm:"not null;comment:时间戳"`
	Line      *int64 `json:"line" gorm:"not null;comment:行号"`
	Content   string `json:"content" gorm:"comment:内容"`
}
