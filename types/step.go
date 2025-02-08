package types

type SStepRes struct {
	Name    string   `json:"name" yaml:"name"`
	State   string   `json:"state" yaml:"state"`
	Code    int64    `json:"code" yaml:"code"`
	Desc    string   `json:"desc,omitempty" yaml:"desc,omitempty"`
	Timeout string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Disable bool     `json:"disable,omitempty" yaml:"disable,omitempty"`
	Depends []string `json:"depends,omitempty" yaml:"depends,omitempty"`
	Message string   `json:"message" yaml:"message"`
	Env     SEnvs    `json:"env,omitempty" yaml:"env,omitempty"`
	Type    string   `json:"type,omitempty" yaml:"type,omitempty"`
	Content string   `json:"content,omitempty" yaml:"content,omitempty"`
	Action  string   `json:"action,omitempty" yaml:"action,omitempty"`
	Rule    string   `json:"rule,omitempty" yaml:"rule,omitempty"`
	Time    STimeRes `json:"time,omitempty" yaml:"time,omitempty"`
}

type SStepsRes []*SStepRes

type SStepReq struct {
	Name    string   `json:"name,omitempty" form:"name" yaml:"name,omitempty"`
	Desc    string   `json:"desc,omitempty" form:"desc" yaml:"desc,omitempty"`
	Timeout string   `json:"timeout,omitempty" form:"timeout" yaml:"timeout,omitempty"`
	Disable bool     `json:"disable,omitempty" form:"disable" yaml:"disable,omitempty"`
	Depends []string `json:"depends,omitempty" form:"depends" yaml:"depends,omitempty"`
	Env     SEnvs    `json:"env,omitempty" form:"env" yaml:"env,omitempty"`
	Type    string   `json:"type,omitempty" form:"type" yaml:"type,omitempty" binding:"required"`
	Content string   `json:"content,omitempty" form:"content" yaml:"content,omitempty" binding:"required"`
	Action  string   `json:"action,omitempty" form:"action" yaml:"action,omitempty"`
	Rule    string   `json:"rule,omitempty" form:"rule" yaml:"rule,omitempty"`
}

type SStepsReq []*SStepReq

type SStepLogRes struct {
	Timestamp int64  `json:"timestamp" yaml:"timestamp"`
	Line      int64  `json:"line" yaml:"line"`
	Content   string `json:"content" yaml:"content"`
}

type SStepLogsRes []*SStepLogRes
