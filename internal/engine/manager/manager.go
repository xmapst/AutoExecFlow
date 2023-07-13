package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var manager = sync.Map{}

const (
	taskPrefix = "task"
	stepPrefix = "step"
)

type tracker struct {
	Name       string
	Context    context.Context
	CancelFunc context.CancelFunc
}

func newTracker(name string, ctx context.Context) tracker {
	t := tracker{
		Name: name,
	}
	t.Context, t.CancelFunc = context.WithCancel(ctx)
	return t
}

func leave(key string) error {
	value, ok := manager.Load(key)
	if !ok {
		return errors.New("not found")
	}
	t, ok := value.(tracker)
	if !ok {
		return errors.New("close exception")
	}
	t.CancelFunc()
	return nil
}

func join(ctx context.Context, key string) context.Context {
	t := newTracker(key, ctx)
	manager.Store(key, t)
	return t.Context
}

func TaskRunning(task string) bool {
	key := fmt.Sprintf("%s#%s", taskPrefix, task)
	_, ok := manager.Load(key)
	return ok
}

func AddTask(ctx context.Context, task string) context.Context {
	key := fmt.Sprintf("%s#%s", taskPrefix, task)
	return join(ctx, key)
}

func AddTaskStep(ctx context.Context, task string, step int64) context.Context {
	key := fmt.Sprintf("%s#%s#%d", stepPrefix, task, step)
	return join(ctx, key)
}

func CloseTask(task string) error {
	key := fmt.Sprintf("%s#%s", taskPrefix, task)
	return leave(key)
}

func CloseTaskStep(task string, step int64) error {
	key := fmt.Sprintf("%s#%s#%d", stepPrefix, task, step)
	return leave(key)
}

func LeaveTask(task string) {
	key := fmt.Sprintf("%s#%s", taskPrefix, task)
	manager.Delete(key)
}

func LeaveTaskStep(task string, step int64) {
	key := fmt.Sprintf("%s#%s#%d", stepPrefix, task, step)
	manager.Delete(key)
}
