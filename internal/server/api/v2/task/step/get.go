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
// @Summary		Detail
// @description	Get step detail
// @Tags		Step
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Param		step path string true "step name"
// @Success		200 {object} types.Base[types.TaskStepRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v2/task/{task}/step/{step} [get]
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
