package router

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/xmapst/osreapi/internal/docs"
	"github.com/xmapst/osreapi/internal/router/api"
	"github.com/xmapst/osreapi/internal/router/api/v1/pool"
	"github.com/xmapst/osreapi/internal/router/api/v1/task"
	"github.com/xmapst/osreapi/internal/router/api/v1/task/step"
	"github.com/xmapst/osreapi/internal/router/api/v1/task/workspace"
	taskv2 "github.com/xmapst/osreapi/internal/router/api/v2/task"
	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/middleware/limiter"
	"github.com/xmapst/osreapi/internal/router/middleware/zap"
	"github.com/xmapst/osreapi/internal/router/types"
)

// @title           OS Remote Executor API
// @version         1.0
// @description     Operating system remote execution interface.

// @contact.name   osreapi
// @contact.url    https://github.com/xmapst/osreapi/issues

// @license.name  GPL-3.0
// @license.url   https://github.com/xmapst/osreapi/blob/main/LICENSE

func New(maxRequests int64) *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.New()
	router.Use(
		cors.Default(),
		gzip.Gzip(gzip.DefaultCompression),
		func(c *gin.Context) {
			c.Header("Server", "Gin")
			c.Header("X-Server", "Gin")
			c.Header("X-Powered-By", "XMapst")
			c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
			c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
			c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			c.Header("Pragma", "no-cache")
		},
		zap.Logger,
		zap.Recovery,
	)

	// debug pprof
	pprof.Register(router)

	// swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// base
	router.GET("/version", version)
	router.GET("/healthyz", healthyz)
	router.GET("/metrics", metrics)
	router.GET("/heartbeat", heartbeat)
	router.HEAD("/heartbeat", heartbeat)
	apiGroup := router.Group("/api", limiter.New(maxRequests, http.MethodPost))
	// V1
	{
		// task
		apiGroup.GET("/v1/task", task.List)
		apiGroup.POST("/v1/task", task.Post)
		apiGroup.GET("/v1/task/:task", task.Get)
		apiGroup.PUT("/v1/task/:task", task.Manager)
		// workspace
		apiGroup.GET("/v1/task/:task/workspace", workspace.Get)
		apiGroup.DELETE("/v1/task/:task/workspace", workspace.Delete)
		apiGroup.POST("/v1/task/:task/workspace", workspace.Post)
		// step
		apiGroup.GET("/v1/task/:task/step/:step", step.Get)
		apiGroup.PUT("/v1/task/:task/step/:step", step.Manager)
		// worker pool
		apiGroup.GET("/v1/pool", pool.Detail)
		apiGroup.POST("/v1/pool", pool.Post)
	}
	// V2
	{
		// task
		apiGroup.POST("/v2/task", taskv2.Post)
	}

	// pty
	router.GET("/api/pty", api.PtyWs)

	// endpoints
	router.Any("/api/endpoints", func(c *gin.Context) {
		render := base.Gin{Context: c}
		var res []types.Endpoint
		var scheme = "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		for _, v := range router.Routes() {
			res = append(res, types.Endpoint{
				Method: v.Method,
				Path:   fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, v.Path),
			})
		}
		render.SetRes(res)
	})

	// no method
	router.NoMethod(func(c *gin.Context) {
		render := base.Gin{Context: c}
		render.SetError(base.CodeErrNoData, errors.New("method not allowed"))
	})

	// no route
	router.NoRoute(func(c *gin.Context) {
		render := base.Gin{Context: c}
		render.SetError(base.CodeErrNoData, errors.New("the requested path does not exist"))
	})
	return router
}
