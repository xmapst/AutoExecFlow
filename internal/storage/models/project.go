package models

type Project struct {
	Base
	Name string `json:"name,omitempty" gorm:"index;not null;comment:名称"`
	ProjectUpdate
}

type ProjectUpdate struct {
	Description string `json:"description,omitempty" gorm:"comment:描述"`
	Disable     *bool  `json:"disable,omitempty" gorm:"not null;default:false;comment:禁用"`
	Type        string `json:"type,omitempty" gorm:"index;not null;comment:类型"`
	Content     string `json:"content,omitempty" gorm:"type:text;comment:内容"`
}

type Projects []*Project
