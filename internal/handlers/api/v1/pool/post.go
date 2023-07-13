package pool

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/logx"
)

type Setting struct {
	Size int `json:"size" form:"command_type" binding:"required" example:"30"`
}

// Post
// @Summary pool setting
// @description post task step
// @Tags Pool
// @Accept json
// @Produce json
// @Param setting body Setting true "pool setting"
// @Success 200 {object} base.Result
// @Failure 500 {object} base.Result
// @Router /api/v1/pool [post]
func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	var req Setting
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
