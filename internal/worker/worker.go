package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/queues"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/event"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var (
	pool        = tunny.NewCallback(1)
	taskManager sync.Map
	stepManager sync.Map
)

func Start(ctx context.Context) error {
	if err := queues.SubscribeTask(ctx, config.App.NodeName, func(data string) error {
		if data == "" {
			return errors.New("invalid task name")
		}
		t, err := newTask(data)
		if err != nil {
			return err
		}
		return pool.Submit(t.Execute)
	}); err != nil {
		return err
	}

	if err := queues.SubscribeManager(ctx, config.App.NodeName, func(data string) error {
		if !utils.ContainsInvisibleChar(data) {
			return errors.New("invalid manager operate")
		}
		slice := utils.SplitByInvisibleChar(data)
		switch len(slice) {
		case 3:
			taskName := slice[0]
			action := slice[1]
			duration := slice[2]
			return managerTask(taskName, action, duration)
		case 4:
			taskName := slice[0]
			stepName := slice[1]
			action := slice[2]
			duration := slice[3]
			return managerStep(taskName, stepName, action, duration)
		}

		return nil
	}); err != nil {
		return err
	}
	_event, id, err := event.SubscribeEvent()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				event.UnSubscribeEvent(id)
				return
			case e := <-_event:
				_ = queues.PublishEvent(e)
			}
		}
	}()
	return nil
}

func managerTask(taskName, action, duration string) error {
	t, err := storage.Task(taskName).Get()
	if err != nil {
		return err
	}

	value, ok := taskManager.Load(taskName)
	if !ok {
		return errors.New("task not found")
	}
	task, ok := value.(*sTask)
	switch action {
	case "kill":
		task.Stop()
		return storage.Task(taskName).Update(&models.STaskUpdate{
			State:    models.Pointer(models.StateFailed),
			OldState: t.State,
			Message:  "has been killed",
		})
	case "pause":
		if *t.State == models.StateRunning {
			return errors.New("step is running")
		}
		if atomic.CompareAndSwapInt32(&task.state, 0, 1) {
			var d time.Duration
			d, err = time.ParseDuration(duration)
			if err == nil && d > 0 {
				task.ctrlCtx, task.ctrlCancel = context.WithTimeout(context.Background(), d)
			} else {
				task.ctrlCtx, task.ctrlCancel = context.WithCancel(context.Background())
			}
			return storage.Task(taskName).Update(&models.STaskUpdate{
				State:    models.Pointer(models.StatePaused),
				OldState: t.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if atomic.CompareAndSwapInt32(&task.state, 1, 0) {
			if task.ctrlCancel != nil {
				task.ctrlCancel()
			}
			return storage.Task(taskName).Update(&models.STaskUpdate{
				State:    t.OldState,
				OldState: t.State,
				Message:  "has been resumed",
			})
		}
	}
	return nil
}

func managerStep(taskName, stepName, action, duration string) error {
	value, ok := stepManager.Load(fmt.Sprintf("%s/%s", taskName, stepName))
	if !ok {
		return errors.New("step not found")
	}
	step, ok := value.(*sStep)
	if !ok {
		return errors.New("step not found")
	}
	s, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		return err
	}
	switch action {
	case "kill":
		step.Stop()
		return storage.Task(taskName).Step(stepName).Update(&models.SStepUpdate{
			State:    models.Pointer(models.StateFailed),
			OldState: s.State,
			Message:  "has been killed",
		})
	case "pause":
		if *s.State == models.StateRunning {
			return errors.New("step is running")
		}
		if atomic.CompareAndSwapInt32(&step.state, 0, 1) {
			var d time.Duration
			d, err = time.ParseDuration(duration)
			if err == nil && d > 0 {
				step.ctrlCtx, step.ctrlCancel = context.WithTimeout(context.Background(), d)
			} else {
				step.ctrlCtx, step.ctrlCancel = context.WithCancel(context.Background())
			}
			return storage.Task(taskName).Step(stepName).Update(&models.SStepUpdate{
				State:    models.Pointer(models.StatePaused),
				OldState: s.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if atomic.CompareAndSwapInt32(&step.state, 1, 0) {
			if step.ctrlCancel != nil {
				step.ctrlCancel()
			}
			return storage.Task(taskName).Step(stepName).Update(&models.SStepUpdate{
				State:    s.OldState,
				OldState: s.State,
				Message:  "has been resumed",
			})
		}
	}
	return nil
}

func SetSize(n int) {
	pool.SetSize(n)
}

func GetSize() int {
	return pool.GetSize()
}

func Shutdown() {
	pool.Close()
}
