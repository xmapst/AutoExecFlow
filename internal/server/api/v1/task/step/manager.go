package step

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
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
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	stepName := c.Param("step")
	if stepName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	action := c.DefaultQuery("action", "paused")
	duration := c.DefaultQuery("duration", "-1")
	err := service.Step(taskName, stepName).Manager(action, duration)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithCode[any](types.CodeSuccess))
}
