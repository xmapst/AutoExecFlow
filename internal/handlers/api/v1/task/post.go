package task

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/handlers/types"
	"github.com/xmapst/osreapi/internal/logx"
)

// Post
// @Summary post task
// @description post task step
// @Tags Task
// @Accept json
// @Produce json
// @param id query string false "task id"
// @Param ansync query bool false "task asynchronously" default(false)
// @Param timeout query string false "task timeout"
// @Param scripts body types.Steps true "scripts"
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task [post]
func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	var req = new(types.Task)
	if err := c.ShouldBindQuery(req); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}

	if err := c.ShouldBind(&req.Step); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}

	if err := req.Check(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}

	var task = &cache.Task{
		ID:       req.ID,
		Timeout:  req.TimeoutDuration,
		MetaData: req.Step.GetMetaData(),
	}

	for _, v := range req.Step {
		task.Steps = append(task.Steps, &cache.TaskStep{
			Name:           v.Name,
			CommandType:    v.CommandType,
			CommandContent: v.CommandContent,
			EnvVars:        v.EnvVars,
			DependsOn:      v.DependsOn,
			Timeout:        v.TimeoutDuration,
		})
	}

	// 提交任务
	if err := worker.Submit(task); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}

	c.Request.Header.Set(types.XTaskID, task.ID)
	c.Writer.Header().Set(types.XTaskID, task.ID)
	c.Set(types.XTaskID, task.ID)

	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	render.SetJson(gin.H{
		"count":     len(req.Step),
		"id":        task.ID,
		"url":       fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), task.ID),
		"timestamp": time.Now().UnixNano(),
	})
}
