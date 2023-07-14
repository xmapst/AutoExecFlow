package dag

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type nodeFunc = func(context.Context) error

type Graph struct {
	ctx    context.Context
	cancel context.CancelFunc
	roots  []*node
	mu     sync.Mutex
	wg     *sync.WaitGroup
}

func newGraph(roots []*node) *Graph {
	return &Graph{roots: roots, wg: new(sync.WaitGroup)}
}

type node struct {
	name    string   // 名称
	adjs    []*node  // adjacent 邻接或相邻节点
	ndeps   int      // number of dependencies 节点的依赖数量
	fn      nodeFunc // 顶点函数
	visited bool     // 追踪节点的遍历状态，以防止重复访问或陷入无限循环
}

// 处理当前节点及其邻接节点的任务逻辑。首先处理当前节点的任务函数，并根据返回的错误情况进行相应的处理。
// 然后，遍历邻接节点，对于每个邻接节点，更新其依赖数量，如果依赖数量为 0，则启动一个新的 Goroutine 来处理该邻接节点。
// 这样可以实现图算法中节点之间的并发处理和依赖关系的维护。
func (g *Graph) runNode(n *node, ch chan<- error) {
	defer g.wg.Done()
	err := n.fn(g.ctx)
	if err != nil {
		ch <- err
		return
	}
	for _, _node := range n.adjs {
		select {
		case <-g.ctx.Done():
			break
		default:
			adj := _node
			dec := func() {
				g.mu.Lock()
				defer g.mu.Unlock()
				adj.ndeps--
			}
			dec()
			if adj.ndeps == 0 {
				g.wg.Add(1)
				go g.runNode(adj, ch)
			}
		}
	}
}

// 控制整个图算法的执行流程。它启动一个 Goroutine 来监听错误通道，同时遍历图的根节点并启动相应的 Goroutine 来处理每个根节点。
// 然后等待所有节点的处理完成，最后检查并打印可能发生的错误信息。同时，通过上下文的取消来通知所有 Goroutine 停止处理。
func (g *Graph) run() (result error) {
	chError := make(chan error)
	go func() {
		for err := range chError {
			result = multierror.Append(result, err)
		}
	}()

	for _, _node := range g.roots {
		root := _node
		g.wg.Add(1)
		go g.runNode(root, chError)
	}
	g.wg.Wait()
	close(chError)
	return
}
