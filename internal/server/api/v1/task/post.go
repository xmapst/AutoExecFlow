package task

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Post
// @Summary		Create
// @description	Create a task
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Accept		multipart/form-data
// @Produce		application/json
// @Produce		application/yaml
// @param		name query string false "task name"
// @Param		async query bool false "task asynchronously" default(false)
// @Param		timeout query string false "task timeout"
// @Param		env query []string false "task envs"
// @Param		steps body types.TaskStepsReq true "scripts"
// @Success		200 {object} types.Base[types.TaskCreateRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task [post]
func Post(c *gin.Context) {
	var req = new(types.TaskReq)
	if err := c.ShouldBindQuery(req); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	if err := c.ShouldBind(&req.Step); err != nil {
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
