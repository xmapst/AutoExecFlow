package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/server/api/middleware/zap"
	"github.com/xmapst/AutoExecFlow/internal/server/api/v1/event"
	"github.com/xmapst/AutoExecFlow/internal/server/api/v1/pool"
	"github.com/xmapst/AutoExecFlow/internal/server/api/v1/sys"
	"github.com/xmapst/AutoExecFlow/internal/server/api/v1/task"
	"github.com/xmapst/AutoExecFlow/internal/server/api/v1/task/step"
	"github.com/xmapst/AutoExecFlow/internal/server/api/v1/task/workspace"
	taskv2 "github.com/xmapst/AutoExecFlow/internal/server/api/v2/task"
	stepv2 "github.com/xmapst/AutoExecFlow/internal/server/api/v2/task/step"
	"github.com/xmapst/AutoExecFlow/pkg/info"
	"github.com/xmapst/AutoExecFlow/types"
)

// New
// @title			Auto Exec Flow
// @version			1.0
// @Description		An `API` for cross-platform custom orchestration of execution steps
// @Description		without any third-party dependencies. Based on `DAG`, it implements the scheduling
// @Description		function of sequential execution of dependent steps and concurrent execution of
// @Description		non-dependent steps. <br /><br /> It provides `API` remote operation mode, batch
// @Description		execution of `Shell` , `Powershell` , `Python` and other commands, and easily
// @Description		completes common management tasks such as running automated operation and maintenance
// @Description		scripts, polling processes, installing or uninstalling software, updating applications,
// @Description		and installing patches.
// @contact.name	AutoExecFlow
// @contact.url		https://github.com/xmapst/AutoExecFlow/issues
// @license.name	GPL-3.0
// @license.url		https://github.com/xmapst/AutoExecFlow/blob/main/LICENSE
func New(relativePath string) *gin.Engine {
	router := gin.New()
	router.Use(
		zap.Logger,
		zap.Recovery,
		cors.Default(),
		gzip.Gzip(gzip.DefaultCompression),
		func(c *gin.Context) {
			c.Header("Server", "Gin")
			c.Header("X-Server", "Gin")
			c.Header("X-Version", info.Version)
			c.Header("X-Powered-By", info.UserEmail)
		},
	)
	baseGroup := router.Group(relativePath)
	// debug pprof
	pprof.Register(baseGroup)
	// base
	baseGroup.GET("/version", version)
	baseGroup.GET("/healthyz", healthyz)
	baseGroup.GET("/heartbeat", heartbeat)
	baseGroup.HEAD("/heartbeat", heartbeat)
	api := baseGroup.Group("/api")
	// V1
	{
		// event
		api.GET("/v1/event", event.Stream)
		// task
		api.GET("/v1/task", task.List)
		api.POST("/v1/task", task.Post)
		api.PUT("/v1/task/:task", task.Manager)
		api.DELETE("/v1/task/:task", task.Delete)
		// dump
		api.GET("/v1/task/:task/dump", task.Dump)
		// workspace
		api.GET("/v1/task/:task/workspace", workspace.Get)
		api.DELETE("/v1/task/:task/workspace", workspace.Delete)
		api.POST("/v1/task/:task/workspace", workspace.Post)
		// step
		api.GET("/v1/task/:task", step.List)
		api.GET("/v1/task/:task/step/:step", step.Log)
		api.PUT("/v1/task/:task/step/:step", step.Manager)
		// worker pool
		api.GET("/v1/pool", pool.Detail)
		api.POST("/v1/pool", pool.Post)

		// pty
		api.GET("/v1/pty", sys.PtyWs)
	}
	// V2
	{
		// task
		api.POST("/v2/task", taskv2.Post)
		api.GET("/v2/task/:task", taskv2.Detail)
		api.GET("/v2/task/:task/step/:step", stepv2.Detail)
	}

	// no method
	router.NoMethod(func(c *gin.Context) {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("method not allowed")))
	})

	// no route
	router.NoRoute(staticHandler(relativePath))
	return router
}
