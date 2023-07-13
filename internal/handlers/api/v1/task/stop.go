package task

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/engine/manager"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/logx"
)

// Stop
// @Summary task stop
// @description stop task
// @Tags Task
// @Accept json
// @Produce json
// @Param task path string true "task id"
// @Success 200 {object} base.Result
// @Failure 500 {object} base.Result
// @Router /api/v1/task/{task} [put]
func Stop(c *gin.Context) {
	render := base.Gin{Context: c}
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeErrParam, errors.New("task does not exist"))
		return
	}
	if err := manager.CloseTask(task); err != nil {
		render.SetError(base.CodeErrNoData, err)
		return
	}
	render.SetJson(nil)
}

// StopStep
// @Summary task step stop
// @description stop task step
// @Tags Task
// @Accept json
// @Produce json
// @Param task path string true "task id"
// @Param step path string true "step id" default(0)
// @Success 200 {object} base.Result
// @Failure 500 {object} base.Result
// @Router /api/v1/task/{task}/{step} [put]
func StopStep(c *gin.Context) {
	render := base.Gin{Context: c}
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeErrParam, errors.New("task does not exist"))
		return
	}
	step, err := strconv.ParseInt(c.Param("step"), 10, 64)
	if err != nil {
		logx.Warningln(err)
		render.SetError(base.CodeErrParam, errors.New("step does not exist"))
		return
	}
	if err := manager.CloseTaskStep(task, step); err != nil {
		render.SetError(base.CodeErrNoData, err)
		return
	}
	render.SetJson(nil)
}
