package types

const (
	XTaskName  = "X-Task-Name"
	XTaskState = "X-Task-STATE"
)

// Response struct

type TaskCreateRes struct {
	// 任务名称
	Name string `json:"name" yaml:"Name"`
	// 步骤数量
	Count int `json:"count" yaml:"Count"`
	// Deprecated, use Name
	ID string `json:"id" yaml:"ID" swaggerignore:"true"`
}

type TaskListRes struct {
	// 分页
	Page PageRes `json:"page" yaml:"Page"`
	// 任务列表
	Tasks []*TaskRes `json:"tasks" yaml:"Tasks"`
}

type PageRes struct {
	// 当前页
	Current int64 `json:"current" yaml:"Current"`
	// 页大小
	Size int64 `json:"size" yaml:"Size"`
	// 总页数
	Total int64 `json:"total" yaml:"Total"`
}

type TaskRes struct {
	// 任务名称
	Name string `json:"name" yaml:"Name"`
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
	Time *TimeRes `json:"time,omitempty" yaml:"Time,omitempty"`
}

type TaskStepRes struct {
	// 步骤名称
	Name string `json:"name" yaml:"Name"`
	// 步骤退出码
	Code int64 `json:"code" yaml:"Code"`
	// 步骤状态
	State string `json:"state" yaml:"State"`
	// 步骤信息
	Message string `json:"msg" yaml:"Message"`
	// 步骤超时时间
	Timeout string `json:"timeout,omitempty" yaml:"Timeout,omitempty"`
	// 是否禁用
	Disable bool `json:"disable" yaml:"Disable"`
	// 步骤依赖
	Depends []string `json:"depends,omitempty" yaml:"Depends,omitempty"`
	// 步骤环境变量
	Env map[string]string `json:"env,omitempty" yaml:"Env,omitempty"`
	// 步骤类型
	Type string `json:"type,omitempty" yaml:"Type,omitempty"`
	// 步骤内容
	Content string `json:"content,omitempty" yaml:"Content,omitempty"`
	// 时间
	Time *TimeRes `json:"time,omitempty" yaml:"Time,omitempty"`
}

type TimeRes struct {
	// 开始时间
	Start string `json:"start,omitempty" yaml:"Start,omitempty"`
	// 结束时间
	End string `json:"end,omitempty" yaml:"End,omitempty"`
}

type TaskStepLogRes struct {
	// 时间戳
	Timestamp int64 `json:"timestamp" yaml:"Timestamp"`
	// 行号
	Line int64 `json:"line" yaml:"Line"`
	// 内容
	Content string `json:"content" yaml:"Content"`
}

// Request struct

type TaskStepReq struct {
	// 步骤名称
	Name string `json:"name,omitempty" form:"name" yaml:"Name,omitempty" example:"script.ps1"`
	// 步骤超时时间
	Timeout string `json:"timeout,omitempty" form:"timeout" yaml:"Timeout,omitempty" example:"3m"`
	// 是否禁用
	Disable bool `json:"disable,omitempty" form:"disable" yaml:"Disable,omitempty" example:"false"`
	// 步骤依赖
	Depends []string `json:"depends,omitempty" form:"depends" yaml:"Depends,omitempty" example:""`
	// 步骤环境变量
	Env map[string]string `json:"env,omitempty" form:"env" yaml:"Env,omitempty" example:"key:value,key1:value1"`
	// 步骤超时时间
	Type string `json:"type,omitempty" form:"type" yaml:"Type,omitempty" example:"powershell"`
	// 步骤内容
	Content string `json:"content,omitempty" form:"content" yaml:"Content,omitempty" example:"sleep 10"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars,omitempty" form:"env_vars" yaml:"EnvVars,omitempty" example:"env1=value1,env2=value2" swaggerignore:"true"`
	// Deprecated, use Type
	CommandType string `json:"command_type,omitempty" form:"command_type" yaml:"CommandType,omitempty" example:"powershell" swaggerignore:"true"`
	// Deprecated, use Content
	CommandContent string `json:"command_content,omitempty" form:"command_content" yaml:"CommandContent,omitempty" example:"sleep 10" swaggerignore:"true"`
}

type TaskReq struct {
	// 任务名称
	Name string `json:"name,omitempty" query:"name"  form:"name" yaml:"Name,omitempty" example:"task_name"`
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
	Step TaskStepsReq `json:"step,omitempty" form:"step" yaml:"Step,omitempty"`

	// Deprecated, use Env
	EnvVars []string `json:"env_vars,omitempty" query:"env_vars" form:"env_vars" yaml:"EnvVars,omitempty" example:"" swaggerignore:"true"`
}

type TaskStepsReq []*TaskStepReq

type PageReq struct {
	Page   int64  `json:"page" query:"page" yaml:"Page" example:"1"`
	Size   int64  `json:"size" query:"size" yaml:"Size" example:"10"`
	Prefix string `json:"prefix" query:"prefix" yaml:"Prefix" example:""`
}
