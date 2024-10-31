package types

type STaskCreateRes struct {
	Name  string `json:"name" yaml:"Name"`
	Count int64  `json:"count,omitempty" yaml:"Count,omitempty"`
}

type STaskListDetailRes struct {
	Page  SPageRes  `json:"page" yaml:"Page"`
	Tasks STasksRes `json:"tasks" yaml:"Tasks"`
}

type STasksRes []*STaskRes

type STaskRes struct {
	Count   int64    `json:"count" yaml:"Count"`
	Desc    string   `json:"desc,omitempty" yaml:"Desc,omitempty"`
	Name    string   `json:"name" yaml:"Name"`
	Node    string   `json:"node,omitempty" yaml:"Node,omitempty"`
	State   string   `json:"state" yaml:"State"`
	Message string   `json:"message" yaml:"Message"`
	Env     SEnvs    `json:"env,omitempty" yaml:"Env,omitempty"`
	Timeout string   `json:"timeout,omitempty" yaml:"Timeout,omitempty"`
	Disable bool     `json:"disable,omitempty" yaml:"Disable,omitempty"`
	Time    STimeRes `json:"time,omitempty" yaml:"Time,omitempty"`
}

type STaskReq struct {
	Name    string    `json:"name,omitempty" form:"name" yaml:"Name,omitempty"`
	Desc    string    `json:"desc,omitempty" yaml:"Desc,omitempty"`
	Node    string    `json:"node,omitempty" form:"node" yaml:"Node,omitempty"`
	Async   bool      `json:"async,omitempty" form:"async" yaml:"Async,omitempty"`
	Disable bool      `json:"disable,omitempty" form:"disable" yaml:"Disable,omitempty"`
	Timeout string    `json:"timeout,omitempty" form:"timeout,omitempty" yaml:"Timeout,omitempty"`
	Env     SEnvs     `json:"env,omitempty" form:"env" yaml:"Env,omitempty"`
	Step    SStepsReq `json:"step,omitempty" form:"step" yaml:"Step,omitempty" binding:"required"`
}
