package project

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary 	更新
// @Description 更新项目
// @Tags 		项目
// @Accept		application/json
// @Produce		application/json
// @Param		project path string true "项目名称"
// @Param		content body types.SProjectUpdateReq true "更新内容"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/project/{project} [post]
func Update(c *gin.Context) {
	projectName := c.Param("project")
	if projectName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	var req = new(types.SProjectUpdateReq)
	if err := c.ShouldBind(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	if err := service.Project(projectName).Update(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
}
