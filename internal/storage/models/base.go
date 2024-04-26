package models

const (
	Stop    int = iota // 成功
	Running            // 运行
	Failed             // 失败
	Unknown            // 未知
	Pending            // 等待
	Paused             // 挂起
)

var StateMap = map[int]string{
	Stop:    "stopped",
	Running: "running",
	Failed:  "failed",
	Unknown: "unknown",
	Pending: "pending",
	Paused:  "paused",
}

type Env struct {
	Name  string `json:"name" gorm:"not null;comment:名称"`
	Value string `json:"value" gorm:"comment:值"`
}

func Pointer[T any](v T) *T {
	return &v
}
