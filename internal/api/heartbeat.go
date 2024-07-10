package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// heartbeat
// @Summary		Heartbeat
// @description	heartbeat
// @Tags		Default
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} any
// @Failure		500 {object} any
// @Router		/heartbeat [get]
func heartbeat(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}
