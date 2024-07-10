package router

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
)

// healthyz
// @Summary		Healthyz
// @description	healthyz
// @Tags		Default
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.Base[types.Healthyz]
// @Failure		500 {object} types.Base[any]
// @Router		/healthyz [get]
func healthyz(c *gin.Context) {
	base.Send(c, types.WithData(&types.Healthyz{
		Server: c.Request.Host,
		Client: c.Request.RemoteAddr,
		State:  "Running",
	}))
}
