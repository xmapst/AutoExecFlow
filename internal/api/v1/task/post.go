package task

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
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
// @Param		steps body types.TaskSteps true "scripts"
// @Success		200 {object} types.Base[types.TaskCreateRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task [post]
func Post(c *gin.Context) {
	var req = new(types.Task)
	if err := c.ShouldBindQuery(req); err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	if err := c.ShouldBind(&req.Step); err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	if err := req.Save(); err != nil {
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(err))
		return
	}

	c.Request.Header.Set(types.XTaskName, req.Name)
	c.Header(types.XTaskName, req.Name)

	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if c.GetHeader("X-Forwarded-Proto") != "" {
		scheme = c.GetHeader("X-Forwarded-Proto")
	}

	base.Send(c, types.WithData(&types.TaskCreateRes{
		URL:   fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), req.Name),
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	}))
}
