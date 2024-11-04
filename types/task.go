package types

type STaskCreateRes struct {
	Name  string `json:"name" yaml:"name"`
	Count int64  `json:"count,omitempty" yaml:"count,omitempty"`
}

type STaskListDetailRes struct {
	Page  SPageRes  `json:"page" yaml:"page"`
	Tasks STasksRes `json:"tasks" yaml:"tasks"`
}

type STasksRes []*STaskRes

type STaskRes struct {
	Count   int64    `json:"count,omitempty" yaml:"count,omitempty"`
	Desc    string   `json:"desc,omitempty" yaml:"desc,omitempty"`
	Name    string   `json:"name" yaml:"name"`
	Node    string   `json:"node,omitempty" yaml:"node,omitempty"`
	State   string   `json:"state" yaml:"state"`
	Message string   `json:"message" yaml:"message"`
	Env     SEnvs    `json:"env,omitempty" yaml:"env,omitempty"`
	Timeout string   `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Disable bool     `json:"disable,omitempty" yaml:"disable,omitempty"`
	Time    STimeRes `json:"time,omitempty" yaml:"time,omitempty"`
}

type STaskReq struct {
	Name    string    `json:"name,omitempty" form:"name" yaml:"name,omitempty"`
	Desc    string    `json:"desc,omitempty" form:"desc" yaml:"desc,omitempty"`
	Node    string    `json:"node,omitempty" form:"node" yaml:"node,omitempty"`
	Async   bool      `json:"async,omitempty" form:"async" yaml:"async,omitempty"`
	Disable bool      `json:"disable,omitempty" form:"disable" yaml:"disable,omitempty"`
	Timeout string    `json:"timeout,omitempty" form:"timeout,omitempty" yaml:"timeout,omitempty"`
	Env     SEnvs     `json:"env,omitempty" form:"env" yaml:"env,omitempty"`
	Step    SStepsReq `json:"step,omitempty" form:"step" yaml:"step,omitempty" binding:"required"`
}
