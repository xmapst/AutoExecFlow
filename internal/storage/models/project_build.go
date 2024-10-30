package models

type SProjectBuild struct {
	SBase
	ProjectName string `json:"project_name,omitempty" gorm:"size:256;uniqueIndex:idx_project_task_name;not null;comment:项目名称"`
	TaskName    string `json:"task_name,omitempty" gorm:"size:256;uniqueIndex:idx_project_task_name;not null;comment:任务名称"`
	Params      string `json:"params,omitempty" gorm:"comment:参数"`
}

func (s *SProjectBuild) TableName() string {
	return "t_project_build"
}

type SProjectBuilds []*SProjectBuild
