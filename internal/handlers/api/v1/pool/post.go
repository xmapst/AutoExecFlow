package pool

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/handlers/types"
	"github.com/xmapst/osreapi/internal/logx"
)

// Post
// @Summary pool setting
// @description post task step
// @Tags Pool
// @Accept json
// @Produce json
// @Param setting body types.PoolSetting true "pool setting"
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/pool [post]
func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	var req types.PoolSetting
	if err := c.ShouldBind(&req); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}
	worker.SetSize(req.Size)
	render.SetJson(nil)
}
