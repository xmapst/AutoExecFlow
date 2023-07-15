package zap

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xmapst/osreapi/internal/logx"
)

func Logger(c *gin.Context) {
	start := time.Now().UTC()
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	c.Next()
	end := time.Now().UTC()
	latency := end.Sub(start)

	fields := []zapcore.Field{
		zap.Int("status", c.Writer.Status()),
		zap.String("protocol", c.Request.Proto),
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.String("ip", c.ClientIP()),
		zap.String("user-agent", c.Request.UserAgent()),
		zap.Duration("latency", latency),
	}
	if query != "" {
		fields = append(fields, zap.String("query", query))
	}

	if len(c.Errors) > 0 {
		for _, e := range c.Errors.Errors() {
			logx.Errorx(e, fields...)
		}
	} else {
		logx.Infox("access path "+path, fields...)
	}
}

func Recovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			// Check for a broken connection, as it is not really a
			// condition that warrants a panic stack trace.
			var brokenPipe bool
			if ne, ok := err.(*net.OpError); ok {
				if se, ok := ne.Err.(*os.SyscallError); ok {
					if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
						brokenPipe = true
					}
				}
			}

			httpRequest, _ := httputil.DumpRequest(c.Request, false)
			if brokenPipe {
				logx.Errorx(c.Request.URL.Path,
					zap.Any("error", err),
					zap.String("request", string(httpRequest)),
				)
				// If the connection is dead, we can't write a status to it.
				_ = c.Error(err.(error)) // nolint: errcheck
				c.Abort()
				return
			}

			logx.Errorx("[Recovery from panic]",
				zap.Time("time", time.Now()),
				zap.Any("error", err),
				zap.String("request", string(httpRequest)),
				zap.String("stack", string(debug.Stack())),
			)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}()
	c.Next()
}
