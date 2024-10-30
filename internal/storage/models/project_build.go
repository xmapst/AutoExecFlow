package models

type ProjectBuild struct {
	Base
	ProjectName string `json:"project_name,omitempty" gorm:"index;not null;comment:项目名称"`
	TaskName    string `json:"task_name,omitempty" gorm:"index;not null;comment:任务名称"`
	Params      string `json:"params,omitempty" gorm:"comment:参数"`
}

type ProjectBuilds []*ProjectBuild
