package task

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary		创建
// @Description	创建任务
// @Tags		任务
// @Accept		application/json
// @Produce		application/json
// @Param		task body types.STaskReq true "任务内容"
// @Success		200 {object} types.SBase[types.STaskCreateRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v2/task [post]
func Post(c *gin.Context) {
	var req = new(types.STaskReq)
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

	base.Send(c, base.WithData(&types.STaskCreateRes{
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	}))
}
