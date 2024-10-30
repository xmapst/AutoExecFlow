package pool

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
)

// Detail
// @Summary		Detail
// @Description	Get task pool details
// @Tags		Pool
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.SBase[types.SPool]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pool [get]
func Detail(c *gin.Context) {
	base.Send(c, base.WithData(service.Pool().Get()))
}
