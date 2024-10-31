package types

type SStepRes struct {
	Name    string   `json:"name" yaml:"Name"`
	Desc    string   `json:"desc,omitempty" yaml:"Desc,omitempty"`
	Code    int64    `json:"code" yaml:"Code"`
	State   string   `json:"state" yaml:"State"`
	Message string   `json:"message" yaml:"Message"`
	Timeout string   `json:"timeout,omitempty" yaml:"Timeout,omitempty"`
	Disable bool     `json:"disable,omitempty" yaml:"Disable,omitempty"`
	Depends []string `json:"depends,omitempty" yaml:"Depends,omitempty"`
	Env     SEnvs    `json:"env,omitempty" yaml:"Env,omitempty"`
	Type    string   `json:"type,omitempty" yaml:"Type,omitempty"`
	Content string   `json:"content,omitempty" yaml:"Content,omitempty"`
	Time    STimeRes `json:"time,omitempty" yaml:"Time,omitempty"`
}

type SStepsRes []*SStepRes

type SStepReq struct {
	Name    string   `json:"name,omitempty" form:"name" yaml:"Name,omitempty"`
	Desc    string   `json:"desc,omitempty" form:"desc" yaml:"Desc,omitempty"`
	Timeout string   `json:"timeout,omitempty" form:"timeout" yaml:"Timeout,omitempty"`
	Disable bool     `json:"disable,omitempty" form:"disable" yaml:"Disable,omitempty"`
	Depends []string `json:"depends,omitempty" form:"depends" yaml:"Depends,omitempty"`
	Env     SEnvs    `json:"env,omitempty" form:"env" yaml:"Env,omitempty"`
	Type    string   `json:"type,omitempty" form:"type" yaml:"Type,omitempty" binding:"required"`
	Content string   `json:"content,omitempty" form:"content" yaml:"Content,omitempty" binding:"required"`
}

type SStepsReq []*SStepReq

type SStepLogRes struct {
	Timestamp int64  `json:"timestamp" yaml:"Timestamp"`
	Line      int64  `json:"line" yaml:"Line"`
	Content   string `json:"content" yaml:"Content"`
}

type SStepLogsRes []*SStepLogRes
