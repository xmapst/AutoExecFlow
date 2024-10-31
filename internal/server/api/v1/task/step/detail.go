package step

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/types"
)

// Detail
// @Summary		详情
// @Description	获取步骤详情
// @Tags		步骤
// @Accept		application/json
// @Produce		application/json
// @Param		task path string true "任务名称"
// @Param		step path string true "步骤名称"
// @Success		200 {object} types.SBase[types.SStepRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/task/{task}/step/{step} [get]
func Detail(c *gin.Context) {
	var taskName = c.Param("task")
	if taskName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	code, step, err := service.Step(taskName, stepName).Detail()
	if err != nil {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}

	base.Send(c, base.WithData(step).WithCode(code).WithError(fmt.Errorf(step.Message)))
}
