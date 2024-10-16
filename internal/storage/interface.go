package storage

import (
	"time"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

const All = ""

type IStorage interface {
	Name() string
	Close() (err error)

	Task(name string) ITask
	// TaskCreate 创建任务
	TaskCreate(task *models.Task) (err error)
	// TaskCount 指定状态任务总数, -1为所有
	TaskCount(state models.State) (res int64)
	// TaskList 获取任务,支持分页, 模糊匹配
	TaskList(page, pageSize int64, str string) (res models.Tasks, total int64)
}

type IBase interface {
	// Name 名称
	Name() string
	// ClearAll 清理
	ClearAll() error
	// Remove 删除
	Remove() (err error)
	// State 获取状态
	State() (state models.State, err error)
	// IsDisable 是否禁用
	IsDisable() bool
	// Env 环境变量接口
	Env() IEnv
}

type ITask interface {
	IBase

	// Timeout 超时时间
	Timeout() (res time.Duration, err error)
	// Get 根据名称获取指定任务
	Get() (res *models.Task, err error)
	// Update 更新
	Update(value *models.TaskUpdate) (err error)

	// Step 步骤接口
	Step(name string) IStep
	// StepCreate 创建步骤
	StepCreate(step *models.Step) (err error)
	// StepCount 任务总数
	StepCount() (res int64)
	// StepNameList 所有步骤名称
	StepNameList(str string) (res []string)
	// StepStateList 获取任务下所有步骤状态
	StepStateList(str string) (res map[string]models.State)
	// StepList 获取任务下所有步骤
	StepList(str string) (res models.Steps)
}

type IStep interface {
	IBase
	// TaskName 任务名称
	TaskName() string
	// Timeout 超时时间
	Timeout() (res time.Duration, err error)
	// Type 类型
	Type() (res string, err error)
	// Content 内容
	Content() (res string, err error)
	// Get 根据名称获取指定步骤
	Get() (res *models.Step, err error)
	// Update 更新
	Update(value *models.StepUpdate) (err error)
	// GlobalEnv 全局环境变量接口
	GlobalEnv() IEnv
	// Depend 依赖接口
	Depend() IDepend
	// Log 日志接口
	Log() ILog
}

type ILog interface {
	// List 获取指定任务指定步骤所有日志, 增量查询
	List(latestLine *int64) (res models.StepLogs)
	// Insert 插入
	Insert(log *models.StepLog) (err error)
	Write(content string)
	Writef(format string, args ...interface{})
	RemoveAll() (err error)
}

type IEnv interface {
	List() (res models.Envs)
	Insert(env ...*models.Env) (err error)
	Get(name string) (string, error)
	Remove(name string) (err error)
	RemoveAll() (err error)
}

type IDepend interface {
	List() (res []string)
	Insert(depends ...string) (err error)
	RemoveAll() (err error)
}
