package dag

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Dagcuter struct {
	Tasks          map[string]Task
	executionOrder []string
	mu             sync.Mutex
}

func New(tasks map[string]Task) (*Dagcuter, error) {
	if HasCycle(tasks) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	return &Dagcuter{
		Tasks: tasks,
	}, nil
}

func (d *Dagcuter) Execute(ctx context.Context) (map[string]map[string]any, error) {
	inDegrees := map[string]int{}
	dependents := map[string][]string{}
	results := map[string]map[string]any{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for name, task := range d.Tasks {
		inDegrees[name] = len(task.Dependencies())
		for _, dep := range task.Dependencies() {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	var runTask func(name string)
	runTask = func(name string) {
		defer wg.Done()
		task := d.Tasks[name]

		mu.Lock()
		inputs := make(map[string]any)
		for _, dep := range task.Dependencies() {
			for k, v := range results[dep] {
				inputs[dep+"."+k] = v
			}
		}
		mu.Unlock()

		err := task.PreExecution(ctx, inputs)
		if err != nil {
			select {
			case errCh <- fmt.Errorf("pre execution task %s failed: %w", name, err):
			default:
			}
			return
		}
		output, err := task.Execute(ctx, inputs)
		if err != nil {
			select {
			case errCh <- fmt.Errorf("task %s failed: %w", name, err):
			default:
			}
			return
		}
		err = task.PostExecution(ctx, output)

		if err != nil {
			select {
			case errCh <- fmt.Errorf("post execution task %s failed: %w", name, err):
			default:
			}
			return
		}
		d.mu.Lock()
		d.executionOrder = append(d.executionOrder, name)
		d.mu.Unlock()
		mu.Lock()
		results[name] = output
		for _, child := range dependents[name] {
			inDegrees[child]--
			if inDegrees[child] == 0 {
				wg.Add(1)
				go runTask(child)
			}
		}
		mu.Unlock()
	}

	for name, deg := range inDegrees {
		if deg == 0 {
			wg.Add(1)
			go runTask(name)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return results, nil
	case err := <-errCh:
		return nil, err
	}
}

func (d *Dagcuter) ExecutionOrder() string {
	var sb = strings.Builder{}
	sb.WriteString("\n")
	for i, step := range d.executionOrder {
		_, _ = fmt.Fprintf(&sb, "%d. %s\n", i+1, step)
	}
	return sb.String()
}
