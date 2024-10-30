package pool

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary		Setting
// @Description	Configuring the Task Pool Size
// @Tags		Pool
// @Accept		application/json
// @Accept		application/yaml
// @Accept		multipart/form-data
// @Produce		application/json
// @Produce		application/yaml
// @Param		setting body types.SPool true "pool setting"
// @Success		200 {object} types.SBase[types.SPool]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pool [post]
func Post(c *gin.Context) {
	var req = new(types.SPool)
	if err := c.ShouldBind(&req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	res, err := service.Pool().Set(req.Size)
	if err != nil {
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithData(res))
}
