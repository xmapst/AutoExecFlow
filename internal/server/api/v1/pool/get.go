package pool

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	_ "github.com/xmapst/AutoExecFlow/types"
)

// Detail
// @Summary		详情
// @Description	获取工作池信息
// @Tags		工作池
// @Accept		application/json
// @Produce		application/json
// @Success		200 {object} types.SBase[types.SPoolRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pool [get]
func Detail(c *gin.Context) {
	base.Send(c, base.WithData(service.Pool().Get()))
}
