package task

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary		CreateV2
// @description	Create a task
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Accept		multipart/form-data
// @Produce		application/json
// @Produce		application/yaml
// @Param		task body types.TaskReq true "scripts"
// @Success		200 {object} types.Base[types.TaskCreateRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v2/task [post]
func Post(c *gin.Context) {
	var req = new(types.TaskReq)
	if err := c.ShouldBind(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	if err := service.Task(req.Name).Create(req); err != nil {
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	c.Request.Header.Set(types.XTaskName, req.Name)
	c.Header(types.XTaskName, req.Name)

	base.Send(c, base.WithData(&types.TaskCreateRes{
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	}))
}
