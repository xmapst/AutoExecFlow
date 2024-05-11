package step

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
	stepName := chi.URLParam(r, "step")
	if stepName == "" {
		base.SendJson(w, base.New().WithData(base.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	action := r.URL.Query().Get("action")
	if action == "" {
		action = "paused"
	}
	duration := r.URL.Query().Get("duration")
	manager, err := dag.VertexManager(taskName, stepName)
	if err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
		return
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			err = storage.Task(taskName).Step(stepName).Update(&models.StepUpdate{
				State:    models.Pointer(models.Failed),
				OldState: step.State,
				Message:  "has been killed",
			})
		}
	case "pause":
		if *step.State == models.Running {
			base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(dag.ErrRunning))
			return
		}
		if manager.State() != dag.Paused {
			_ = manager.Pause(duration)
			err = storage.Task(taskName).Step(stepName).Update(&models.StepUpdate{
				State:    models.Pointer(models.Paused),
				OldState: step.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.State() == dag.Paused {
			manager.Resume()
			err = storage.Task(taskName).Step(stepName).Update(&models.StepUpdate{
				State:    step.OldState,
				OldState: step.State,
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
