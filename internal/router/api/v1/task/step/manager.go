package step

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Manager(c *gin.Context) {
	render := base.Gin{Context: c}
	taskName := c.Param("task")
	if taskName == "" {
		render.SetError(base.CodeNoData, errors.New("task does not exist"))
		return
	}
	stepName := c.Param("step")
	if stepName == "" {
		render.SetError(base.CodeNoData, errors.New("step does not exist"))
		return
	}
	action := c.DefaultQuery("action", "paused")
	duration := c.Query("duration")
	manager, err := dag.VertexManager(taskName, stepName)
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeNoData, err)
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeNoData, err)
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
			render.SetError(base.CodeFailed, dag.ErrRunning)
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
		render.SetError(base.CodeFailed, err)
		return
	}
	render.SetRes(nil)
}
