package types

type SPipelineListRes struct {
	Page      SPageRes      `json:"page" yaml:"Page"`
	Pipelines SPipelinesRes `json:"pipelines" yaml:"Pipelines"`
}

type SPipelinesRes []*SPipelineRes

type SPipelineRes struct {
	Name    string `json:"name" yaml:"Name"`
	Desc    string `json:"desc,omitempty" yaml:"Desc,omitempty"`
	Disable bool   `json:"disable,omitempty" yaml:"Disable,omitempty"`
	TplType string `json:"tpl_type" yaml:"TplType"`
	Content string `json:"content,omitempty" yaml:"Content,omitempty"`
}

type SPipelineCreateReq struct {
	Name    string `json:"name" yaml:"Name" binding:"required"`
	Desc    string `json:"desc,omitempty" yaml:"Desc,omitempty"`
	Disable *bool  `json:"disable" yaml:"Disable"`
	TplType string `json:"tpl_type" yaml:"TplType" binding:"required" example:"jinja2"`
	Content string `json:"content" yaml:"Content" binding:"required"`
}

type SPipelineUpdateReq struct {
	Desc    string `json:"desc,omitempty" yaml:"Desc,omitempty"`
	Disable *bool  `json:"disable" yaml:"Disable"`
	TplType string `json:"tpl_type" yaml:"TplType" binding:"required" example:"jinja2"`
	Content string `json:"content" yaml:"Content" binding:"required"`
}

type SPipelineBuildRes struct {
	Pipeline string `json:"pipeline" yaml:"Pipeline"`
	TaskName string `json:"task_name" yaml:"TaskName"`
	Params   string `json:"params,omitempty" yaml:"Params,omitempty"`
}

type SPipelineBuildReq struct {
	Params map[string]any `json:"params,omitempty" yaml:"Params,omitempty"`
}

type SPipelineBuildListRes struct {
	Page  SPageRes `json:"page" yaml:"Page"`
	Tasks []string `json:"tasks" yaml:"Tasks"`
}
