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

// Manager
// @Summary task step manager
// @description manager task step
// @Tags Task
// @Accept application/json
// @Accept application/toml
// @Accept application/x-yaml
// @Accept multipart/form-data
// @Produce application/json
// @Produce application/x-yaml
// @Produce application/toml
// @Param task path string true "task name"
// @Param step path string true "step name"
// @Param action query string false "management action" Enums(paused,kill,pause,resume) default(paused)
// @Param duration query string false "how long to pause; if empty, manual continuation is required" default(1m)
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task/{task}/step/{step} [put]
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
		if !manager.Paused() {
			_ = manager.Pause(duration)
			err = storage.Task(taskName).Step(stepName).Update(&models.StepUpdate{
				State:    models.Pointer(models.Paused),
				OldState: step.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.Paused() {
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
