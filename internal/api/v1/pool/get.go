package pool

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/worker"
)

// Detail
// @Summary		Detail
// @description	Get task pool details
// @Tags		Pool
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.Base[types.Pool]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/pool [get]
func Detail(c *gin.Context) {
	base.Send(c, types.WithData(&types.Pool{
		Size:    worker.GetSize(),
		Total:   worker.GetTotal(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	}),
	)
}
