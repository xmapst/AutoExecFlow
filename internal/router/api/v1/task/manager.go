package task

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Manager(w http.ResponseWriter, r *http.Request) {
	taskName := chi.URLParam(r, "task")
	if taskName == "" {
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	action := base.DefaultQuery(r, "action", "paused")
	duration := base.Query(r, "duration")
	manager, err := dag.GraphManager(taskName)
	if err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
		return
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
		return
	}
	if *task.State <= models.Stop || *task.State >= models.Failed {
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(errors.New("task is no running")))
		return
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			err = storage.Task(taskName).Update(&models.TaskUpdate{
				State:    models.Pointer(models.Failed),
				OldState: task.State,
				Message:  "has been killed",
			})
		}
	case "pause":
		if manager.State() != dag.Paused {
			_ = manager.Pause(duration)
			err = storage.Task(taskName).Update(&models.TaskUpdate{
				State:    models.Pointer(models.Paused),
				OldState: task.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.State() == dag.Paused {
			manager.Resume()
			err = storage.Task(taskName).Update(&models.TaskUpdate{
				State:    task.OldState,
				OldState: task.State,
				Message:  "has been resumed",
			})
		}
	}
	if err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(err))
		return
	}
	base.SendJson(w, base.New())
}
