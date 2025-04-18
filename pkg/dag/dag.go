package dag

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Dagcuter struct {
	Tasks          map[string]Task
	results        *sync.Map
	inDegrees      map[string]int
	dependents     map[string][]string
	executionOrder []string
	mu             *sync.Mutex
	wg             *sync.WaitGroup
}

func New(tasks map[string]Task) (*Dagcuter, error) {
	if HasCycle(tasks) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	return &Dagcuter{
		mu:         new(sync.Mutex),
		wg:         new(sync.WaitGroup),
		results:    new(sync.Map),
		inDegrees:  make(map[string]int),
		dependents: make(map[string][]string),
		Tasks:      tasks,
	}, nil
}

func (d *Dagcuter) Execute(ctx context.Context) (map[string]map[string]any, error) {
	defer d.results.Clear()
	errCh := make(chan error, 1)

	for name, task := range d.Tasks {
		d.inDegrees[name] = len(task.Dependencies())
		for _, dep := range task.Dependencies() {
			d.dependents[dep] = append(d.dependents[dep], name)
		}
	}

	for name, deg := range d.inDegrees {
		if deg == 0 {
			d.wg.Add(1)
			go d.runTask(ctx, name, errCh)
		}
	}

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		results := make(map[string]map[string]any)
		d.results.Range(func(key, value any) bool {
			results[key.(string)] = value.(map[string]any)
			return true
		})
		return results, nil
	case err := <-errCh:
		return nil, err
	}
}

func (d *Dagcuter) runTask(ctx context.Context, name string, errCh chan error) {
	defer d.wg.Done()
	task := d.Tasks[name]

	d.mu.Lock()
	inputs := d.prepareInputs(task)
	d.mu.Unlock()

	output, err := d.executeTask(ctx, name, task, inputs)
	if err != nil {
		select {
		case errCh <- fmt.Errorf("task %s failed: %w", name, err):
		default:
		}
		return
	}

	d.mu.Lock()
	d.executionOrder = append(d.executionOrder, name)
	d.mu.Unlock()

	d.mu.Lock()
	d.results.Store(name, output)
	for _, child := range d.dependents[name] {
		d.inDegrees[child]--
		if d.inDegrees[child] == 0 {
			d.wg.Add(1)
			go d.runTask(ctx, child, errCh)
		}
	}
	d.mu.Unlock()
}

func (d *Dagcuter) executeTask(ctx context.Context, name string, task Task, inputs map[string]any) (map[string]any, error) {
	if err := task.PreExecution(ctx, inputs); err != nil {
		return nil, fmt.Errorf("pre execution task %s failed: %w", name, err)
	}
	output, err := task.Execute(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("task %s failed: %w", name, err)
	}
	if err = task.PostExecution(ctx, output); err != nil {
		return nil, fmt.Errorf("post execution task %s failed: %w", name, err)
	}
	return output, nil
}

func (d *Dagcuter) prepareInputs(task Task) map[string]any {
	inputs := make(map[string]any)
	for _, dep := range task.Dependencies() {
		d.results.Range(func(key, value any) bool {
			if key.(string) == dep {
				inputs[dep] = value
			}
			return true
		})
	}
	return inputs
}

func (d *Dagcuter) ExecutionOrder() string {
	var sb = strings.Builder{}
	sb.WriteString("\n")
	for i, step := range d.executionOrder {
		_, _ = fmt.Fprintf(&sb, "%d. %s\n", i+1, step)
	}
	return sb.String()
}
