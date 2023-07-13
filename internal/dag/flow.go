package dag

import (
	"context"
	"errors"
)

var (
	ErrCycleDetected = errors.New("dependency cycle detected")
	ErrEmptyTask     = errors.New("empty Task cannot run")
)

type Task struct {
	steps []*Step
}

func NewTask() *Task {
	return &Task{}
}

// Add 用于向 Flow 实例添加任务。如果任务已经有一个非零的 ID，表示该任务已经在 Flow 中存在，直接返回该任务。
// 否则，给任务分配一个唯一的 ID，并将任务添加到 Flow 的任务列表中。
func (t *Task) Add(s *Step) *Step {
	if s.id > 0 {
		return t.steps[s.id-1]
	}

	s.id = len(t.steps) + 1
	s.t = t

	t.steps = append(t.steps, s)
	return s
}

// Compile 将任务列表转换为有向无环图。它遍历任务列表中的每个任务，创建对应的节点，并建立节点之间的依赖关系。
// 在建立依赖关系时，将依赖的节点的指针添加到当前节点的 adjs 切片中。
// 如果检测到回环，则返回 ErrCycleDetected 错误。如果图中没有根节点，则返回 ErrEmptyFlow 错误。
// 最后，通过调用 newGraph 函数创建一个新的图对象，并将根节点传递给它。
func (t *Task) Compile() (*Graph, error) {
	// 存储图中的根节点。
	var roots []*node
	// 根据任务的数量创建了一个切片 nodes，切片的长度与任务数量相同。
	nodes := make([]*node, len(t.steps))

	// 循环，为每个节点初始化相关属性，并将节点的指针存储在 nodes 切片中。
	// 每个节点都有一个 adjs 切片，用于存储与之相邻的节点。
	for i := 0; i < len(nodes); i++ {
		nodes[i] = new(node)
		nodes[i].adjs = make([]*node, 0)
	}

	// 循环遍历任务列表 f.tasks，为每个节点设置相应的属性，并根据任务的依赖关系将节点连接起来。
	for i, task := range t.steps {
		nodes[i].name = task.name
		nodes[i].fn = task.fn
		nodes[i].ndeps = len(task.deps)

		// 将当前顶点作为邻接或相邻分配给父顶点。
		// 具体地, 将依赖的节点的指针添加到当前节点的 adjs 切片中，表示当前节点依赖于这些节点。
		for _, dep := range task.deps {
			p := nodes[dep.id-1]
			p.adjs = append(p.adjs, nodes[i])
		}
		// 如果任务没有依赖，将其节点添加到 roots 切片中。
		if len(task.deps) == 0 {
			roots = append(roots, nodes[i])
		}

		// 检查图表是否存在回环, 检查以本节点为起点的子图是否存在回环。
		if err := t.detectCircularDependencies(nodes[i], []*node{}); err != nil {
			return nil, err
		}
	}
	if len(roots) == 0 {
		return nil, ErrEmptyTask
	}
	return newGraph(roots), nil
}

// 使用深度优先搜索 (DFS) 的方式进行回环检测。它从给定的节点开始遍历邻接节点，并在遍历过程中检查是否存在回环。
// 如果发现已访问过的节点，则存在回环，返回 ErrCycleDetected 错误。否则，继续递归遍历邻接节点。
// 为了避免重复访问节点，使用 visited 属性对已访问的节点进行标记。
func (t *Task) detectCircularDependencies(current *node, path []*node) error {
	// 如果发现某个邻接节点已经被访问过（即存在回环）
	if current.visited {
		return ErrCycleDetected
	}
	current.visited = true
	// 递归地遍历节点的邻接节点
	for _, adj := range current.adjs {
		if err := t.detectCircularDependencies(adj, append(path, current)); err != nil {
			return err
		}
	}

	current.visited = false
	return nil
}

// Run 运行任务流程。它首先调用 compile 方法生成有向无环图，并检查是否有错误发生。
// 如果有错误，则打印错误信息并返回。否则，调用图对象的 run 方法运行任务流程。
func (t *Task) Run(ctx context.Context) error {
	// 生成有向无环图
	g, err := t.Compile()
	if err != nil {
		return err
	}
	g.ctx, g.cancel = context.WithCancel(ctx)
	defer g.cancel()
	return g.run()
}

type StepFunc = func(context.Context) error

type Step struct {
	id   int      // 步骤id
	name string   // 步骤名称
	deps []*Step  // 依赖列表
	fn   StepFunc // 步骤函数
	t    *Task
}

// NewStep 创建一个新的步骤实例
func NewStep(name string, fn StepFunc) *Step {
	return &Step{name: name, fn: fn}
}

// WithDeps 为任务添加依赖任务。它会检查依赖任务是否已经在 Flow 中存在，如果不存在，则将依赖任务添加到 Flow 中
func (t *Step) WithDeps(steps ...*Step) {
	// adds deps that are not added in flow
	for _, task := range steps {
		if task.id == 0 {
			t.t.Add(task)
		}
	}
	t.deps = append(t.deps, steps...)
}

// Name 获取步骤的名称
func (t *Step) Name() string {
	return t.name
}
