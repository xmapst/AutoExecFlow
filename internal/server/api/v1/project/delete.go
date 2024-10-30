package project

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Delete
// @Summary 	删除项目
// @Description 删除项目
// @Tags 		项目
// @Accept		application/json
// @Produce		application/json
// @Param		project path string true "项目名称"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/project/{project} [delete]
func Delete(c *gin.Context) {
	projectName := c.Param("project")
	if projectName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("project does not exist")))
		return
	}
	err := service.Project(projectName).Delete()
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithCode[any](types.CodeSuccess))
}
