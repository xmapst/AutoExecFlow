package worker

import (
	"context"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/queues"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/tunny"
)

var pool = tunny.NewCallback(1)

func Start(ctx context.Context) error {
	if err := queues.SubscribeTask(ctx, config.App.NodeName, func(data string) error {
		if data == "" {
			return errors.New("invalid task name")
		}
		t, err := newTask(data)
		if err != nil {
			return err
		}
		return pool.Submit(t.run)
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
	event, id, err := dag.SubscribeEvent()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				dag.UnSubscribeEvent(id)
				return
			case e := <-event:
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

	manager, err := dag.GraphManager(taskName)
	if err != nil {
		return err
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			return storage.Task(taskName).Update(&models.STaskUpdate{
				State:    models.Pointer(models.StateFailed),
				OldState: t.State,
				Message:  "has been killed",
			})
		}
	case "pause":
		if manager.State() != dag.StatePaused {
			_ = manager.Pause(duration)
			return storage.Task(taskName).Update(&models.STaskUpdate{
				State:    models.Pointer(models.StatePaused),
				OldState: t.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.State() == dag.StatePaused {
			manager.Resume()
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
	manager, err := dag.VertexManager(taskName, stepName)
	if err != nil {
		return err
	}
	s, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		return err
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			return storage.Task(taskName).Step(stepName).Update(&models.SStepUpdate{
				State:    models.Pointer(models.StateFailed),
				OldState: s.State,
				Message:  "has been killed",
			})
		}
	case "pause":
		if *s.State == models.StateRunning {
			return dag.ErrRunning
		}
		if manager.State() != dag.StatePaused {
			_ = manager.Pause(duration)
			return storage.Task(taskName).Step(stepName).Update(&models.SStepUpdate{
				State:    models.Pointer(models.StatePaused),
				OldState: s.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.State() == dag.StatePaused {
			manager.Resume()
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
