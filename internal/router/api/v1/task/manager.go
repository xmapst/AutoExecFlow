package task

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
)

// Manager
// @Summary task manager
// @description manager task
// @Tags Task
// @Accept application/json
// @Accept application/toml
// @Accept application/x-yaml
// @Accept multipart/form-data
// @Produce application/json
// @Produce application/x-yaml
// @Produce application/toml
// @Param task path string true "task name"
// @Param action query string false "management action" Enums(paused,kill,pause,resume) default(paused)
// @Param duration query string false "how long to pause; if empty, manual continuation is required" default(1m)
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task/{task} [put]
func Manager(c *gin.Context) {
	render := base.Gin{Context: c}
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeErrNoData, errors.New("task does not exist"))
		return
	}
	action := c.DefaultQuery("action", "paused")
	duration := c.Query("duration")
	manager, err := dag.GraphManager(task)
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeErrNoData, err)
		return
	}
	taskDetail, err := storage.TaskDetail(task)
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeErrNoData, err)
		return
	}
	if taskDetail.State <= exec.Stop {
		render.SetError(base.CodeExecErr, errors.New("task is no running"))
		return
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			taskDetail.State = exec.Killed
			err = storage.SetTask(task, taskDetail)
		}
	case "pause":
		if !manager.Paused() {
			_ = manager.Pause(duration)
			taskDetail.OldState, taskDetail.State = taskDetail.State, exec.Paused
			err = storage.SetTask(task, taskDetail)
		}
	case "resume":
		if manager.Paused() {
			manager.Resume()
			taskDetail.State, taskDetail.OldState = taskDetail.OldState, taskDetail.State
			err = storage.SetTask(task, taskDetail)
		}
	}
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeExecErr, err)
		return
	}
	render.SetRes(nil)
}
