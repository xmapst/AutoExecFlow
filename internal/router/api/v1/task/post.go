package task

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	var req = new(types.Task)
	if err := c.ShouldBindQuery(req); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeFailed, err)
		return
	}

	if err := c.ShouldBind(&req.Step); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeFailed, err)
		return
	}

	if err := req.Save(); err != nil {
		render.SetError(base.CodeFailed, err)
		return
	}

	// 提交任务
	if err := worker.Submit(req.Name); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeFailed, err)
		return
	}

	c.Request.Header.Set(types.XTaskName, req.Name)
	c.Writer.Header().Set(types.XTaskName, req.Name)
	c.Set(types.XTaskName, req.Name)

	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	render.SetRes(&types.TaskCreateRes{
		URL:   fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), req.Name),
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	})
}
