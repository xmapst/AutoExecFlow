package task

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
// @description	Task management, can terminate, suspend, and resolve
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Param		action query string false "management action" Enums(paused,kill,pause,resume) default(paused)
// @Param		duration query string false "how long to pause; if empty, manual continuation is required" default(1m)
// @Success		200 {object} types.Base[any]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task/{task} [put]
func Manager(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	action := c.DefaultQuery("action", "paused")
	duration := c.Query("duration")
	manager, err := dag.GraphManager(taskName)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	if *task.State <= models.Stop || *task.State >= models.Failed {
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(errors.New("task is no running")))
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
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, types.WithCode[any](types.CodeSuccess))
}
