package pool

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/handlers/base"
)

// Detail
// @Summary pool detail
// @description detail pool
// @Tags Pool
// @Accept json
// @Produce json
// @Success 200 {object} base.Result
// @Failure 500 {object} base.Result
// @Router /api/v1/pool [get]
func Detail(c *gin.Context) {
	render := base.Gin{Context: c}
	render.SetJson(gin.H{
		"size":      worker.GetSize(),
		"running":   worker.Running(),
		"waiting":   worker.Waiting(),
		"timestamp": time.Now().UnixNano(),
	})
}
