package backend

import (
	"time"

	"github.com/xmapst/osreapi/internal/storage/models"
)

const All = ""

type IStorage interface {
	Name() string
	Close() (err error)

	Task(name string) ITask
	// TaskList 获取所有任务
	TaskList(str string) (res models.Tasks)
}

type IBase interface {
	// Name 名称
	Name() string
	// ClearAll 清理
	ClearAll()
	// Delete 删除
	Delete() (err error)
	// State 获取状态
	State() (state int, err error)

	// Env 环境变量接口
	Env() IEnv
}

type ITask interface {
	IBase

	// Timeout 超时时间
	Timeout() (res time.Duration, err error)
	// Get 根据名称获取指定任务
	Get() (res *models.Task, err error)
	// Create 插入
	Create(task *models.Task) (err error)
	// Update 更新
	Update(value *models.TaskUpdate) (err error)

	// Step 步骤接口
	Step(name string) IStep
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
	// Create 插入
	Create(step *models.Step) (err error)
	// Update 更新
	Update(value *models.StepUpdate) (err error)

	// Depend 依赖接口
	Depend() IDepend
	// Log 日志接口
	Log() ILog
}

type ILog interface {
	// List 获取指定任务指定步骤所有日志,支持分页
	List() (res models.Logs)
	// Create 插入
	Create(log *models.Log) (err error)
}

type IEnv interface {
	List() (res models.Envs)
	Create(env ...*models.Env) (err error)
	Get(name string) (string, error)
	Delete(name string) (err error)
	DeleteAll() (err error)
}

type IDepend interface {
	List() (res []string)
	Create(depends ...string) (err error)
	DeleteAll() (err error)
}
