package pipeline

import (
	"github.com/gin-gonic/gin"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary 	创建
// @Description 创建流水线
// @Tags 		流水线
// @Accept		application/json
// @Produce		application/json
// @Param		content body types.SPipelineCreateReq true "流水线内容"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pipeline [post]
func Post(c *gin.Context) {
	var req = new(types.SPipelineCreateReq)
	if err := c.ShouldBind(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	if err := service.Pipeline(req.Name).Create(req); err != nil {
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	base.Send(c, base.WithCode[any](types.CodeSuccess))
}
