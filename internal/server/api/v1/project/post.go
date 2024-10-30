package project

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary 	创建
// @Description 创建项目
// @Tags 		项目
// @Accept		application/json
// @Produce		application/json
// @Param		content body types.SProjectCreateReq true "项目内容"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/project [post]
func Post(c *gin.Context) {
	var req = new(types.SProjectCreateReq)
	if err := c.ShouldBind(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	if err := service.Project(req.Name).Create(req); err != nil {
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	base.Send(c, base.WithCode[any](types.CodeSuccess))
}
