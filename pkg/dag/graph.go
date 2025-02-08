package dag

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type IGraph interface {
	IManager

	AddVertex(v *Vertex) (*Vertex, error)
	Validator(force bool) error
	Run(ctx context.Context) (err error)
}

type sGraph struct {
	wg  *sync.WaitGroup
	ctx *sContext

	vertex []*Vertex // 顶点集
}

func New(name string) IGraph {
	g := &sGraph{
		wg: new(sync.WaitGroup),
		ctx: &sContext{
			name: name,
		},
	}

	// 基础上下文
	g.ctx.baseCtx, g.ctx.baseCancel = context.WithCancel(context.Background())

	// 加入管理
	join(fmt.Sprintf(graphPrefix, g.Name()), g)

	return g
}

// Name 获取名称
func (g *sGraph) Name() string {
	return g.ctx.name
}

// Kill 强杀
func (g *sGraph) Kill() error {
	if g.ctx.baseCancel == nil {
		return ErrContext
	}

	g.ctx.baseCancel()
	remove(fmt.Sprintf(graphPrefix, g.Name()))
	emitEvent("kill task %s", g.Name())
	return nil
}

// Pause 挂起
func (g *sGraph) Pause(duration string) error {
	g.ctx.Lock()
	defer g.ctx.Unlock()

	if g.ctx.state == StatePaused || g.ctx.controlCtx != nil {
		// 重复挂起, 直接返回
		return nil
	}
	g.ctx.oldState = g.ctx.state
	g.ctx.state = StatePaused

	d, err := time.ParseDuration(duration)
	if err == nil && d > 0 {
		g.ctx.controlCtx, g.ctx.controlCancel = context.WithTimeout(context.Background(), d)
	} else {
		g.ctx.controlCtx, g.ctx.controlCancel = context.WithCancel(context.Background())
	}
	emitEvent("pause task %s", g.Name())

	return nil
}

// Resume 解挂
func (g *sGraph) Resume() {
	g.ctx.Lock()
	defer g.ctx.Unlock()
	if g.ctx.state != StatePaused || g.ctx.controlCancel == nil {
		// 没有挂起不需要恢复,直接返回
		return
	}
	g.ctx.oldState = g.ctx.state
	g.ctx.state = StateResume
	// 解除挂起
	g.ctx.controlCancel()
	emitEvent("resume task %s", g.Name())
}

// WaitResume 等待解挂
func (g *sGraph) WaitResume() {
	g.ctx.Lock()
	defer g.ctx.Unlock()

	if g.ctx.state != StatePaused || g.ctx.controlCtx == nil {
		// 没有挂起不需要d等待,直接返回
		return
	}
	<-g.ctx.controlCtx.Done()
}

// State 返回当前状态
func (g *sGraph) State() State {
	return g.ctx.state
}

// AddVertex 添加顶点
func (g *sGraph) AddVertex(v *Vertex) (*Vertex, error) {
	g.ctx.Lock()
	defer g.ctx.Unlock()

	if g.ctx.visited {
		emitEvent("duplicate step %s in task %s", v.Name(), g.Name())
		return nil, ErrDuplicateCompile
	}

	if v.cid > 0 {
		return g.vertex[v.cid-1], nil
	}

	// 初始化顶点基础上下文
	v.ctx.baseCtx, v.ctx.baseCancel = context.WithCancel(g.ctx.baseCtx)

	// 设置顶点id
	v.cid = int64(len(g.vertex) + 1)

	g.vertex = append(g.vertex, v)
	v.graph = g

	join(fmt.Sprintf(vertexPrefix, g.Name(), v.Name()), v)

	return v, nil
}

// 处理当前节点及其邻接节点的任务逻辑。首先处理当前节点的任务函数，并根据返回的错误情况进行相应的处理。
// 然后，遍历邻接节点，对于每个邻接节点，更新其依赖数量，如果依赖数量为 0，则启动一个新的 Goroutine 来处理该邻接节点。
// 这样可以实现图算法中节点之间的并发处理和依赖关系的维护。
func (g *sGraph) runVertex(v *Vertex, errCh chan<- error) {
	var err error
	defer g.wg.Done()
	emitEvent("start step %s in task %s", v.Name(), g.Name())
	defer func() {
		v.ctx.mainCancel()

		g.ctx.Lock()
		v.ctx.oldState = v.ctx.state
		g.ctx.state = StateStopped
		g.ctx.Unlock()

		remove(fmt.Sprintf(vertexPrefix, g.Name(), v.Name()))
		if err != nil {
			errCh <- err
			emitEvent("error exec step %s in task %s, %s", v.Name(), g.Name(), err)
			return
		}
		emitEvent("stoped step %s in task %s", v.Name(), g.Name())
	}()

	// 图形级暂停
	if g.State() == StatePaused {
		emitEvent("step %s paused because task %s is paused", v.Name(), g.Name())
		select {
		case <-g.ctx.mainCtx.Done():
			// 被终止
			err = ErrForceKill
			return
		case <-g.ctx.controlCtx.Done():
			// 继续
			emitEvent("resumed step %s in task %s", v.Name(), g.Name())
		}
	}

	// 节点级执行前控制, 挂起/解卦/强杀
	if v.State() == StatePaused {
		emitEvent("paused step %s in task %s", v.Name(), g.Name())
		select {
		case <-g.ctx.mainCtx.Done():
			// 被终止
			err = ErrForceKill
			return
		case <-v.ctx.mainCtx.Done():
			// 被终止
			err = ErrForceKill
			return
		case <-v.ctx.controlCtx.Done():
			// 继续
			emitEvent("resumed step %s in task %s", v.Name(), g.Name())
		}
	}

	v.ctx.Lock()
	v.ctx.oldState = v.ctx.state
	v.ctx.state = StateRunning
	v.ctx.Unlock()

	// 执行顶点函数
	if err = v.fn(v.ctx.mainCtx, g.Name(), v.Name()); err != nil {
		return
	}

	// 节点级执行后控制, 挂起/解卦/强杀
	if v.State() == StatePaused {
		emitEvent("paused step %s in task %s", v.Name(), g.Name())
		select {
		case <-v.ctx.controlCtx.Done():
			// 继续
			emitEvent("resumed step %s in task %s", v.Name(), g.Name())
		}
	}

	// 执行后面的顶点
	for k := range v.adjs {
		select {
		case <-g.ctx.mainCtx.Done():
			err = ErrForceKill
			break
		default:
			dec := func() {
				g.ctx.Lock()
				defer g.ctx.Unlock()
				v.adjs[k].ndeps--
			}
			dec()
			if v.adjs[k].ndeps == 0 {
				g.wg.Add(1)
				go g.runVertex(v.adjs[k], errCh)
			}
		}
	}
}

// 合并多个上下文取消
func (g *sGraph) withCancel(main, extra context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(main)
	stopf := context.AfterFunc(extra, func() {
		cancel()
	})
	return ctx, func() {
		cancel()
		stopf()
	}
}

// Run 运行任务流程,控制整个图算法的执行流程。它启动一个 Goroutine 来监听错误通道，同时遍历图的根节点并启动相应的 Goroutine 来处理每个根节点。
// 然后等待所有节点的处理完成，最后检查并打印可能发生的错误信息。同时，通过上下文的取消来通知所有 Goroutine 停止处理。
func (g *sGraph) Run(ctx context.Context) (err error) {
	emitEvent("start task %s", g.Name())
	if err = g.Validator(true); err != nil {
		emitEvent("invalid task %s, %s", g.Name(), err)
		return err
	}
	g.initContext(ctx)

	defer func() {
		g.ctx.mainCancel()
		g.ctx.Lock()
		g.ctx.oldState = g.ctx.state
		g.ctx.state = StateStopped
		g.ctx.Unlock()
		remove(fmt.Sprintf(graphPrefix, g.Name()))
	}()

	var chError = make(chan error, 1)
	var done = make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case _err, ok := <-chError:
				if !ok {
					return
				}
				err = errors.Join(err, _err)
			case <-g.ctx.mainCtx.Done():
				err = ErrForceKill
				emitEvent("task %s ends because task %s is terminated", g.Name(), g.Name())
				return
			}
		}
	}()

	for k := range g.vertex {
		if !g.vertex[k].root {
			continue
		}
		g.wg.Add(1)
		go g.runVertex(g.vertex[k], chError)
	}

	g.wg.Wait()
	close(chError)
	<-done
	if err != nil {
		emitEvent("error exec task %s, %s", g.Name(), err)
		return err
	}
	emitEvent("end task %s", g.Name())
	return
}

func (g *sGraph) initContext(ctx context.Context) {
	// 设置图形主上下文
	g.ctx.mainCtx, g.ctx.mainCancel = g.withCancel(ctx, g.ctx.baseCtx)
	// 设置运行中
	g.ctx.Lock()
	g.ctx.oldState = g.ctx.state
	g.ctx.state = StateRunning
	g.ctx.Unlock()

	// 设置所有顶点主上下文
	for k := range g.vertex {
		g.vertex[k].ctx.mainCtx, g.vertex[k].ctx.mainCancel = g.withCancel(g.ctx.mainCtx, g.vertex[k].ctx.baseCtx)
	}
}

// Validator 检查图的状态，并根据需要进行编译
func (g *sGraph) Validator(force bool) error {
	if g.vertex == nil {
		return ErrEmptyGraph
	}
	if g.ctx.visited && !force {
		// 图已编译且不强制重新编译
		return nil
	}
	return g.compile()
}

// Reset 清除图和顶点的状态，准备重新编译
func (g *sGraph) reset() {
	g.ctx.Lock()
	defer g.ctx.Unlock()

	g.ctx.visited = false // 允许重新编译
	for _, vertex := range g.vertex {
		vertex.ctx.visited = false
		vertex.adjs = []*Vertex{} // 清空邻接节点
		vertex.ndeps = 0          // 重置依赖计数
		vertex.root = false       // 重置根节点标志
	}
}

// compile 将任务列表转换为有向无环图。它遍历任务列表中的每个任务，创建对应的节点，并建立节点之间的依赖关系。
// 在建立依赖关系时，将依赖的节点的指针添加到当前节点的 adjs 切片中。
// 如果检测到回环，则返回 ErrCycleDetected 错误
func (g *sGraph) compile() (err error) {
	// 先重置图的状态，以支持多次编译
	g.reset()

	g.ctx.Lock()
	defer func() {
		// 标记为已编译
		g.ctx.visited = true
		if err != nil {
			// 如果编译失败，恢复未编译状态
			g.ctx.visited = false
			// 清空顶点信息
			g.vertex = nil
		}
		g.ctx.Unlock()
	}()

	var nameMap = make(map[string]bool)

	// 循环遍历任务列表 g.vertex，为每个节点设置相应的属性，并根据任务的依赖关系将节点连接起来。
	for _, v := range g.vertex {
		if _, ok := nameMap[v.Name()]; ok {
			return ErrDuplicateVertexName
		}
		nameMap[v.Name()] = true
		// 设置依赖数量
		v.ndeps = int64(len(v.deps))

		// 将当前顶点作为邻接或相邻分配给父顶点。
		// 具体地, 将依赖的节点的指针添加到当前节点的 adjs 切片中，表示当前节点依赖于这些节点。
		for _, dep := range v.deps {
			g.vertex[dep.cid-1].adjs = append(g.vertex[dep.cid-1].adjs, v)
		}

		// 如果任务没有依赖，将其节点添加到 roots 切片中。
		if len(v.deps) == 0 {
			v.root = true
		}
	}

	// 使用 DFS 进行环检测
	visited := make(map[*Vertex]bool)
	stack := make(map[*Vertex]bool)
	for _, vertex := range g.vertex {
		if err = g.detectCircularDependencies(vertex, visited, stack); err != nil {
			return err
		}
	}

	return
}

// 使用深度优先搜索 (DFS) 的方式进行回环检测。它从给定的节点开始遍历邻接节点，并在遍历过程中检查是否存在回环。
// 如果发现已访问过的节点，则存在回环，返回 ErrCycleDetected 错误。否则，继续递归遍历邻接节点。
// 为了避免重复访问节点，使用 visited 属性对已访问的节点进行标记。
func (g *sGraph) detectCircularDependencies(current *Vertex, visited, stack map[*Vertex]bool) error {
	if stack[current] {
		return ErrCycleDetected
	}
	if visited[current] {
		return nil
	}

	visited[current] = true
	stack[current] = true

	for _, adj := range current.adjs {
		if err := g.detectCircularDependencies(adj, visited, stack); err != nil {
			return err
		}
	}

	stack[current] = false
	return nil
}
