package dag

import (
	"errors"
	"sync"
)

func New() *Runner {
	return new(Runner)
}

// Runner collects functions and arranges them as vertices and edges of a directed acyclic graph.
// Upon validation of the graph, functions are run in parallel topological order. The zero value
// is useful.
type Runner struct {
	fns   map[string]func() error
	graph map[string][]string
	wg    sync.WaitGroup
}

var ErrMissingVertex = errors.New("missing vertex")
var ErrCycleDetected = errors.New("dependency cycle detected")

// AddVertex adds a function as a vertex in the graph. Only functions which have been added in this
// way will be executed during Run.
func (d *Runner) AddVertex(name string, fn func() error) {
	if d.fns == nil {
		d.fns = make(map[string]func() error)
	}
	d.fns[name] = fn
}

// AddEdge establishes a dependency between two vertices in the graph. Both from and to must exist
// in the graph, or Run will err. The vertex at from will execute before the vertex at to.
func (d *Runner) AddEdge(from, to string) {
	if d.graph == nil {
		d.graph = make(map[string][]string)
	}
	d.graph[from] = append(d.graph[from], to)
}

type result struct {
	name string
	err  error
}

func (d *Runner) DetectCycles() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for vertex := range d.graph {
		if !visited[vertex] {
			if d.detectCyclesHelper(vertex, visited, recStack) {
				return true
			}
		}
	}
	return false
}

func (d *Runner) detectCyclesHelper(vertex string, visited, recStack map[string]bool) bool {
	visited[vertex] = true
	recStack[vertex] = true

	for _, v := range d.graph[vertex] {
		// only check cycles on a vertex one time
		if !visited[v] {
			if d.detectCyclesHelper(v, visited, recStack) {
				return true
			}
			// if we've visited this vertex in this recursion stack, then we have a cycle
		} else if recStack[v] {
			return true
		}

	}
	recStack[vertex] = false
	return false
}

// Run will validate that all edges in the graph point to existing vertices, and that there are
// no dependency cycles. After validation, each vertex will be run, deterministically, in parallel
// topological order. If any vertex returns an error, no more vertices will be scheduled and
// Run will exit and return that error once all in-flight functions finish execution.
func (d *Runner) Run() error {
	// sanity check
	if len(d.fns) == 0 {
		return nil
	}
	// count how many deps each vertex has
	deps := make(map[string]int)
	for vertex, edges := range d.graph {
		// every vertex along every edge must have an associated fn
		if _, ok := d.fns[vertex]; !ok {
			return ErrMissingVertex
		}
		for _, edgeVertex := range edges {
			if _, ok := d.fns[edgeVertex]; !ok {
				return ErrMissingVertex
			}
			deps[edgeVertex]++
		}
	}

	if d.DetectCycles() {
		return ErrCycleDetected
	}

	running := 0
	resc := make(chan result, len(d.fns))
	var err error

	// start any vertex that has no deps
	for name := range d.fns {
		if deps[name] == 0 {
			running++
			d.wg.Add(1)
			d.start(name, d.fns[name], resc)
		}
	}

	// wait for all running work to complete
	for running > 0 {
		res := <-resc
		running--

		// capture the first error
		if res.err != nil && err == nil {
			err = res.err
		}

		// don't enqueue any more work on if there's been an error
		if err != nil {
			continue
		}

		// start any vertex whose deps are fully resolved
		for _, vertex := range d.graph[res.name] {
			if deps[vertex]--; deps[vertex] == 0 {
				running++
				d.wg.Add(1)
				d.start(vertex, d.fns[vertex], resc)
			}
		}
	}

	d.wg.Wait()

	return err
}

func (d *Runner) start(name string, fn func() error, resc chan<- result) {
	go func() {
		defer d.wg.Done()
		resc <- result{
			name: name,
			err:  fn(),
		}
	}()
}
