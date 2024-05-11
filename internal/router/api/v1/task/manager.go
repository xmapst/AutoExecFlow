package task

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Manager(w http.ResponseWriter, r *http.Request) {
	taskName := chi.URLParam(r, "task")
	if taskName == "" {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	action := r.URL.Query().Get("action")
	if action == "" {
		action = "paused"
	}
	duration := r.URL.Query().Get("duration")
	manager, err := dag.GraphManager(taskName)
	if err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
		return
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
		return
	}
	if *task.State <= models.Stop || *task.State >= models.Failed {
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(errors.New("task is no running")))
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
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(err))
		return
	}
	render.JSON(w, r, types.New())
}
