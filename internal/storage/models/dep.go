package models

type StepDepend struct {
	Name string `json:"name,omitempty" gorm:"not null,comment:依赖步骤名称"`
}
