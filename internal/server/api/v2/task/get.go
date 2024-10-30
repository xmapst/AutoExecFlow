package task

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
// @Description	Get task detail
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Success		200 {object} types.SBase[types.STaskRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v2/task/{task} [get]
func Detail(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	code, task, err := service.Task(taskName).Detail()
	if err != nil {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	c.Request.Header.Set(types.XTaskState, task.State)
	c.Header(types.XTaskState, task.State)

	base.Send(c, base.WithData(task).WithCode(code).WithError(fmt.Errorf(task.Message)))
}
