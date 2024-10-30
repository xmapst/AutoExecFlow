package api

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/types"
)

// healthyz
// @Summary		Healthyz
// @Description	healthyz
// @Tags		Default
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.SBase[types.SHealthyz]
// @Failure		500 {object} types.SBase[any]
// @Router		/healthyz [get]
func healthyz(c *gin.Context) {
	base.Send(c, base.WithData(&types.SHealthyz{
		Server: c.Request.Host,
		Client: c.Request.RemoteAddr,
		State:  "Running",
	}))
}
