package step

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
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeErrNoData, errors.New("task does not exist"))
		return
	}
	step := c.Param("step")
	if step == "" {
		render.SetError(base.CodeErrNoData, errors.New("step does not exist"))
		return
	}
	action := c.DefaultQuery("action", "paused")
	duration := c.Query("duration")
	manager, err := dag.VertexManager(task, step)
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeErrNoData, err)
		return
	}
	stepDetail, err := storage.TaskStepDetail(task, step)
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeErrNoData, err)
		return
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			stepDetail.State = exec.Killed
			err = storage.SetTaskStep(task, step, stepDetail)
		}
	case "pause":
		if stepDetail.State == exec.Running {
			render.SetError(base.CodeExecErr, dag.ErrRunning)
			return
		}
		if !manager.Paused() {
			_ = manager.Pause(duration)
			stepDetail.State = exec.Paused
			err = storage.SetTaskStep(task, step, stepDetail)
		}
	case "resume":
		if manager.Paused() {
			manager.Resume()
			stepDetail.State = exec.Pending
			err = storage.SetTaskStep(task, step, stepDetail)
		}
	}
	if err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeExecErr, err)
		return
	}
	render.SetRes(nil)
}
