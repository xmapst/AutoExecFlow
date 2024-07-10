package step

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

// Manager
// @Summary		Manager
// @description	Step management, can terminate, suspend, and resolve
// @Tags		Step
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Param		step path string true "step name"
// @Param		action query string false "management action" Enums(paused,kill,pause,resume) default(paused)
// @Param		duration query string false "how long to pause; if empty, manual continuation is required" default(1m)
// @Success		200 {object} types.Base[any]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task/{task}/step/{step} [put]
func Manager(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	stepName := c.Param("step")
	if stepName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	action := c.DefaultQuery("action", "paused")
	duration := c.Query("duration")
	manager, err := dag.VertexManager(taskName, stepName)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
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
			base.Send(c, types.WithCode[any](types.CodeFailed).WithError(dag.ErrRunning))
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
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, types.WithCode[any](types.CodeSuccess))
}
