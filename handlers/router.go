package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xmapst/osreapi/config"
	_ "github.com/xmapst/osreapi/docs"
	"github.com/xmapst/osreapi/engine"
	"math"
	"net/http"
	"time"
)

func Router() *gin.Engine {
	router := gin.New()
	coreConf := cors.DefaultConfig()
	coreConf.AllowAllOrigins = true
	router.Use(gin.Recovery(), cors.New(coreConf), logger())
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
			"Task":   engine.Pool.QueueLength(),
		})
	})
	router.GET("/", List)
	router.POST("/", Post)
	router.GET("/:id", GetTask)
	router.GET("/:id/:step", GetStep)
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

func logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		host := c.Request.Host
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}
		if len(c.Errors) > 0 {
			logrus.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			// client_ip mode path statue_code latency dataLength referer userAgent
			logrus.Infof("%s %s %s %d %d %d %s %s %s", clientIP, c.Request.Method, path, statusCode, latency, dataLength, host, referer, clientUserAgent)
		}
	}
}
