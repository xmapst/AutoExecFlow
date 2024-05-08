package task

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
	action := c.DefaultQuery("action", "paused")
	duration := c.Query("duration")
	manager, err := dag.GraphManager(taskName)
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeNoData, err)
		return
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeNoData, err)
		return
	}
	if *task.State <= models.Stop || *task.State >= models.Failed {
		render.SetError(base.CodeFailed, errors.New("task is no running"))
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
		if !manager.Paused() {
			_ = manager.Pause(duration)
			err = storage.Task(taskName).Update(&models.TaskUpdate{
				State:    models.Pointer(models.Paused),
				OldState: task.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.Paused() {
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
		render.SetError(base.CodeFailed, err)
		return
	}
	render.SetRes(nil)
}
