package models

type State int

const (
	StateStop    State = iota // 成功
	StateRunning              // 运行
	StateFailed               // 失败
	StateUnknown              // 未知
	StatePending              // 等待
	StatePaused               // 挂起
	StateAll     State = -1
)

var StateMap = map[State]string{
	StateStop:    "stopped",
	StateRunning: "running",
	StateFailed:  "failed",
	StateUnknown: "unknown",
	StatePending: "pending",
	StatePaused:  "paused",
}

func Pointer[T any](v T) *T {
	return &v
}
