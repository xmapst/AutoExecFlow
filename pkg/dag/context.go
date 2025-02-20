package dag

import (
	"context"
	"sync"
)

type State int

const (
	StateUnknown State = iota
	StateRunning
	StateStopped
	StatePaused
	StateResume
)

type sContext struct {
	sync.Mutex
	name     string // 名称
	visited  bool   // 编译过或者追踪节点的遍历状态，以防止重复访问或陷入无限循环
	state    State  // 现在状态
	oldState State  // 上一个状态

	// 生命周期控制（强杀）
	lifecycleCtx    context.Context
	lifecycleCancel context.CancelFunc

	// 主执行流程（超时/外部取消）
	executionCtx    context.Context
	executionCancel context.CancelFunc

	// 控制上下文, 控制挂起或解卦
	controlCtx    context.Context
	controlCancel context.CancelFunc
}
