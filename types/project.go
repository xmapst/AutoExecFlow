package types

type SProjectListRes struct {
	Page     SPageRes     `json:"page" yaml:"Page"`
	Projects SProjectsRes `json:"projects" yaml:"Projects"`
}

type SProjectsRes []*SProjectRes

type SProjectRes struct {
	Name        string `json:"name" yaml:"Name"`
	Description string `json:"description,omitempty" yaml:"Description,omitempty"`
	Disable     bool   `json:"disable" yaml:"Disable"`
	Type        string `json:"type" yaml:"Type"`
	Content     string `json:"content" yaml:"Content"`
}

type SProjectCreateReq struct {
	Name string `json:"name" yaml:"Name" binding:"required"`
	SProjectUpdateReq
}

type SProjectUpdateReq struct {
	Description string `json:"description,omitempty" yaml:"Description,omitempty"`
	Disable     *bool  `json:"disable" yaml:"Disable"`
	Type        string `json:"type" yaml:"Type" binding:"required"`
	Content     string `json:"content" yaml:"Content" binding:"required"`
}
