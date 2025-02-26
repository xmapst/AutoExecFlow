package dag

import (
	"context"
	"time"
)

type VertexFunc = func(ctx context.Context, gName, vName string) error

type Vertex struct {
	ctx   *sContext
	graph IGraph

	cid   int64      // 临时id, 从1开始
	fn    VertexFunc // 函数
	adjs  []*Vertex  // 邻接或相邻顶点
	deps  []*Vertex  // 依赖列表
	ndeps int64      // 顶点的依赖数量
	root  bool       // 根顶点
}

func NewVertex(name string, fn VertexFunc) *Vertex {
	v := &Vertex{
		ctx: &sContext{
			name: name,
		},
		fn: fn,
	}
	return v
}

// Name 获取名称
func (v *Vertex) Name() string {
	return v.ctx.name
}

// Kill 强杀
func (v *Vertex) Kill() error {
	if v.ctx.lifecycleCtx == nil {
		return ErrContext
	}

	v.ctx.lifecycleCancel()
	emitEvent("kill step %s in task %s", v.Name(), v.graph.Name())
	return nil
}

// Pause 挂起
func (v *Vertex) Pause(duration string) error {
	v.ctx.Lock()
	defer v.ctx.Unlock()

	if v.ctx.state == StatePaused || v.ctx.controlCtx != nil {
		// 重复挂起, 直接返回
		return nil
	}
	v.ctx.oldState = v.ctx.state
	v.ctx.state = StatePaused

	d, err := time.ParseDuration(duration)
	if err == nil && d > 0 {
		v.ctx.controlCtx, v.ctx.controlCancel = context.WithTimeout(context.Background(), d)
	} else {
		v.ctx.controlCtx, v.ctx.controlCancel = context.WithCancel(context.Background())
	}
	emitEvent("pause step %s in task %s", v.Name(), v.graph.Name())
	return nil
}

// Resume 解挂
func (v *Vertex) Resume() {
	v.ctx.Lock()
	defer v.ctx.Unlock()
	if v.ctx.state != StatePaused || v.ctx.controlCtx == nil {
		// 没有挂起不需要恢复,直接返回
		return
	}
	v.ctx.oldState = v.ctx.state
	v.ctx.state = StateResume
	// 解除挂起
	v.ctx.controlCancel()
	emitEvent("resume step %s in task %s", v.Name(), v.graph.Name())
}

// WaitResume 等待解挂
func (v *Vertex) WaitResume() {
	v.ctx.Lock()
	defer v.ctx.Unlock()

	if v.ctx.state != StatePaused || v.ctx.controlCtx == nil {
		// 没有挂起不需要d等待,直接返回
		return
	}
	<-v.ctx.controlCtx.Done()
}

// State 是否挂起
func (v *Vertex) State() State {
	return v.ctx.state
}

// WithDeps 为顶点添加依赖顶点。它会检查依赖顶点是否已经在图形中存在，如果不存在，则将依赖顶点添加到图型中
func (v *Vertex) WithDeps(vv ...*Vertex) error {
	// adds deps that are not added in graph
	for _, task := range vv {
		if task.cid == 0 {
			if _, err := v.graph.AddVertex(task); err != nil {
				return err
			}
		}
	}
	v.deps = append(v.deps, vv...)
	return nil
}
