package handlers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xmapst/osreapi/internal/config"
	_ "github.com/xmapst/osreapi/internal/docs"
	"github.com/xmapst/osreapi/internal/handlers/api/v1/pool"
	"github.com/xmapst/osreapi/internal/handlers/api/v1/status"
	"github.com/xmapst/osreapi/internal/handlers/api/v1/task"
	"github.com/xmapst/osreapi/internal/handlers/middleware/limiter"
	"github.com/xmapst/osreapi/internal/handlers/middleware/zap"
)

func Router() *gin.Engine {
	router := gin.New()
	coreConf := cors.DefaultConfig()
	coreConf.AllowAllOrigins = true
	router.Use(
		zap.Recovery,
		zap.Logger,
		cors.New(coreConf),
		func(c *gin.Context) {
			c.Header("Server", "Gin")
			c.Header("X-Server", "Gin")
			c.Header("X-Powered-By", "XMapst")
			c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
			c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
			c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			c.Header("Pragma", "no-cache")
			c.Next()
		},
	)
	pprof.Register(router)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/version", version)
	router.GET("/healthyz", healthyz)
	router.GET("/metrics", metrics)
	router.GET("/heartbeat", heartbeat)
	router.HEAD("/heartbeat", heartbeat)
	apiG := router.Group("/api", limiter.New(config.App.MaxRequests, []string{http.MethodPost}))
	{
		v1G := apiG.Group("/v1")
		{
			taskG := v1G.Group("/task")
			{
				taskG.GET("", task.List)
				taskG.POST("", task.Post)
				taskG.GET("/:task", task.Detail)
				taskG.PUT("/:task", task.Stop)
				taskG.PUT("/:task/:step", task.StopStep)
				taskG.GET("/:task/:step/console", task.StepDetail)
			}
			poolG := v1G.Group("/pool")
			{
				poolG.GET("", pool.Detail)
				poolG.POST("", pool.Post)
			}
			stateG := v1G.Group("/state")
			{
				stateG.GET("", status.Detail)
			}
		}
	}

	// Compatible with the original route, it will be deleted in the future
	// Deprecated: Use /api/v1/task
	router.GET("/", task.List)
	// Deprecated: Use /api/v1/task
	router.POST("/", task.Post)
	// Deprecated: Use /api/v1/task/:task
	router.GET("/:task", task.Detail)
	// Deprecated: Use /api/v1/task/:task/:step/console
	router.GET("/:task/:step/console", task.StepDetail)
	return router
}
