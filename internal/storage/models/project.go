package models

type SProject struct {
	SBase
	Name string `json:"name,omitempty" gorm:"index;not null;comment:名称"`
	SProjectUpdate
}

func (s *SProject) TableName() string {
	return "t_project"
}

type SProjectUpdate struct {
	Description string `json:"description,omitempty" gorm:"comment:描述"`
	Disable     *bool  `json:"disable,omitempty" gorm:"not null;default:false;comment:禁用"`
	Type        string `json:"type,omitempty" gorm:"index;not null;comment:类型"`
	Content     string `json:"content,omitempty" gorm:"type:text;comment:内容"`
}

type SProjects []*SProject
