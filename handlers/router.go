package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/juju/ratelimit"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xmapst/osreapi/config"
	_ "github.com/xmapst/osreapi/docs"
	"github.com/xmapst/osreapi/engine"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type accessLog struct {
	TimeStamp  string `json:"timestamp"`
	ClientIP   string `json:"client_ip"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Protocol   string `json:"protocol"`
	StatusCode int    `json:"status"`
	Latency    int64  `json:"duration"`
	BodySize   int    `json:"body_size"`
}

func Router() *gin.Engine {
	router := gin.New()
	coreConf := cors.DefaultConfig()
	coreConf.AllowAllOrigins = true
	router.Use(
		gin.Recovery(), cors.New(coreConf),
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
		gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			log := &accessLog{
				TimeStamp:  param.TimeStamp.Format(time.RFC3339),
				ClientIP:   param.ClientIP,
				Method:     param.Method,
				Path:       param.Path,
				Protocol:   param.Request.Proto,
				StatusCode: param.StatusCode,
				Latency:    int64(param.Latency),
				BodySize:   param.BodySize,
			}
			bs, err := json.Marshal(log)
			if err != nil {
				logrus.Error(err)
				return ""
			}
			// your custom format
			return string(bs) + "\n"
		}),
	)
	lm := newRateLimiter(time.Minute, config.App.MaxRequests, func(ctx *gin.Context) (string, error) {
		key := ctx.Request.Header.Get("X-API-KEY")
		if key != "" {
			return key, nil
		}
		return "", errors.New("API key is missing")
	})
	if config.App.MaxRequests > 0 {
		router.Use(lm.Middleware())
	}
	if config.App.Debug {
		pprof.Register(router)
		// swagger doc
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	router.GET("/version", Version)
	router.GET("/healthyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"Server": c.Request.Host,
			"Client": c.ClientIP(),
			"State":  "Running",
			"Task":   engine.QueueLength(),
		})
	})
	router.GET("/", List)
	router.POST("/", Post)
	router.GET("/:id", GetTask)
	router.GET("/:id/:step/console", GetStep)
	return router
}

// RateKeyFunc limiter
type RateKeyFunc func(ctx *gin.Context) (string, error)

type RateLimiterMiddleware struct {
	fillInterval time.Duration
	capacity     int64
	rateKeyGen   RateKeyFunc
	limiters     map[string]*ratelimit.Bucket
}

func (r *RateLimiterMiddleware) get(ctx *gin.Context) (*ratelimit.Bucket, error) {
	key, err := r.rateKeyGen(ctx)

	if err != nil {
		return nil, err
	}

	if limiter, existed := r.limiters[key]; existed {
		return limiter, nil
	}

	limiter := ratelimit.NewBucketWithQuantum(r.fillInterval, r.capacity, r.capacity)
	r.limiters[key] = limiter
	return limiter, nil
}

func (r *RateLimiterMiddleware) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limiter, err := r.get(ctx)
		if err != nil || limiter.TakeAvailable(1) == 0 {
			if err == nil {
				err = errors.New("too many requests")
			}
			_ = ctx.AbortWithError(429, err)
		} else {
			ctx.Writer.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.Available()))
			ctx.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.Capacity()))
			ctx.Next()
		}
	}
}

func newRateLimiter(interval time.Duration, capacity int64, keyGen RateKeyFunc) *RateLimiterMiddleware {
	limiters := make(map[string]*ratelimit.Bucket)
	return &RateLimiterMiddleware{
		interval,
		capacity,
		keyGen,
		limiters,
	}
}
