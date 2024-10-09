package task

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/types"
)

// Dump
// @Summary		Dump
// @description	dump task
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Success		200 {object} types.Base[types.TaskCreateRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task/{task}/dump [get]

func Dump(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	res, err := service.Task(taskName).Dump()
	if err != nil {
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithData[*types.TaskCreateRes](res))
}
