package taskv2

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

// Post
// @Summary post task
// @description post task step
// @Tags Task
// @Accept application/json
// @Accept application/toml
// @Accept application/x-yaml
// @Accept multipart/form-data
// @Produce application/json
// @Produce application/x-yaml
// @Produce application/toml
// @Param scripts body types.Task true "scripts"
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v2/task [post]
func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	var req = new(types.Task)
	if err := c.ShouldBind(req); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeExecErr, err)
		return
	}

	if err := req.Check(); err != nil {
		render.SetError(base.CodeExecErr, err)
		return
	}

	var task = &worker.Task{
		Name:     req.Name,
		Timeout:  req.TimeoutDuration,
		EnvVars:  req.EnvVars,
		MetaData: req.Step.GetMetaData(),
	}

	for _, v := range req.Step {
		task.Steps = append(task.Steps, &worker.TaskStep{
			Name:           v.Name,
			CommandType:    v.CommandType,
			CommandContent: v.CommandContent,
			EnvVars:        v.EnvVars,
			DependsOn:      v.DependsOn,
			Timeout:        v.TimeoutDuration,
		})
	}

	// 加入池中异步处理
	if err := worker.Submit(task); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeExecErr, err)
		return
	}

	c.Request.Header.Set(types.XTaskName, task.Name)
	c.Writer.Header().Set(types.XTaskName, task.Name)
	c.Set(types.XTaskName, task.Name)

	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	path := strings.Replace(strings.TrimSuffix(c.Request.URL.Path, "/"), "v2", "v1", 1)
	render.SetRes(&types.TaskRes{
		URL:   fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, path, req.Name),
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	})
}
