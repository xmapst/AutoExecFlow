package types

type STaskCreateRes struct {
	// 任务名称
	Name string `json:"name" yaml:"Name"`
	// 步骤数量
	Count int `json:"count" yaml:"Count"`
	// Deprecated, use Name
	ID string `json:"id" yaml:"ID" swaggerignore:"true"`
}

type STaskListDetailRes struct {
	// 分页
	Page SPageRes `json:"page" yaml:"Page"`
	// 任务列表
	Tasks STasksRes `json:"tasks" yaml:"Tasks"`
}

type STasksRes []*STaskRes

type STaskRes struct {
	// 任务名称
	Name string `json:"name" yaml:"Name"`
	// 任务描述
	Description string `json:"description,omitempty" yaml:"Description,omitempty"`
	// 节点名称
	Node string `json:"node,omitempty" yaml:"Node,omitempty"`
	// 任务状态
	State string `json:"state" yaml:"State"`
	// 任务信息
	Message string `json:"msg" yaml:"Message"`
	// 步骤数量
	Count int `json:"count" yaml:"Count"`
	// 任务环境变量
	Env map[string]string `json:"env,omitempty" yaml:"Env,omitempty"`
	// 任务超时时间
	Timeout string `json:"timeout,omitempty" yaml:"Timeout,omitempty"`
	// 是否禁用
	Disable bool `json:"disable" yaml:"Disable"`
	// 时间
	Time *STimeRes `json:"time,omitempty" yaml:"Time,omitempty"`
}

type STaskReq struct {
	// 任务名称
	Name string `json:"name,omitempty" query:"name"  form:"name" yaml:"Name,omitempty" example:"task_name"`
	// 任务描述
	Description string `json:"description,omitempty" query:"description" yaml:"Description,omitempty"`
	// Node 执行节点
	Node string `json:"node,omitempty" query:"node" form:"node" yaml:"Node,omitempty" example:""`
	// 是否异步执行
	Async bool `json:"async,omitempty" query:"async" form:"async" yaml:"Async,omitempty" example:"false"`
	// 是否禁用
	Disable bool `json:"disable,omitempty" query:"disable" form:"disable" yaml:"Disable,omitempty" example:"false"`
	// 任务超时时间
	Timeout string `json:"timeout,omitempty" query:"timeout" form:"timeout,omitempty" yaml:"Timeout,omitempty" example:"24h"`
	// 任务环境变量
	Env map[string]string `json:"env,omitempty" query:"env" form:"env" yaml:"Env,omitempty" example:"key:value,key1:value1"`
	// 任务步骤
	Step SStepsReq `json:"step,omitempty" form:"step" yaml:"Step,omitempty"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars,omitempty" query:"env_vars" form:"env_vars" yaml:"EnvVars,omitempty" example:"" swaggerignore:"true"`
}

type SStepsReq []*SStepReq
