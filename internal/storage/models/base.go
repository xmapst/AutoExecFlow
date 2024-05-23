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

func Pointer[T any](v T) *T {
	return &v
}
