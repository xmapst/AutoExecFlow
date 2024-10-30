package types

type SStepsRes []*SStepRes

type SStepRes struct {
	// 步骤名称
	Name string `json:"name" yaml:"Name"`
	// 步骤描述
	Description string `json:"description,omitempty" yaml:"Description,omitempty"`
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
	Time *STimeRes `json:"time,omitempty" yaml:"Time,omitempty"`
}

type SStepReq struct {
	// 步骤名称
	Name string `json:"name,omitempty" form:"name" yaml:"Name,omitempty" example:"script.ps1"`
	// 步骤描述
	Description string `json:"description,omitempty" form:"description" yaml:"Description,omitempty" example:"script.ps1"`
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

type SLogRes struct {
	// 时间戳
	Timestamp int64 `json:"timestamp" yaml:"Timestamp"`
	// 行号
	Line int64 `json:"line" yaml:"Line"`
	// 内容
	Content string `json:"content" yaml:"Content"`
}

type SLogsRes []*SLogRes
