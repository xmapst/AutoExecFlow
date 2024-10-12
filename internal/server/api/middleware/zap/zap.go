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

	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

func Logger(c *gin.Context) {
	start := time.Now().UTC()
	path := c.Request.URL.String()
	c.Next()
	end := time.Now().UTC()
	latency := end.Sub(start)

	if len(c.Errors) > 0 {
		for _, e := range c.Errors.Errors() {
			logx.Errorln(c.ClientIP(), c.Request.Method, c.Request.Proto, c.Writer.Status(), path, latency, e)
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		logx.Infoln(c.ClientIP(), c.Request.Method, c.Request.Proto, c.Writer.Status(), path, latency, c.Request.UserAgent())
	}
}

func Recovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			handlePanic(c, err)
		}
	}()
	c.Next()
}

func handlePanic(c *gin.Context, err interface{}) {
	// 检查是否是连接中断
	if isBrokenPipeError(err) {
		httpRequest, _ := httputil.DumpRequest(c.Request, false)
		logx.Errorln(c.Request.URL.Path, httpRequest, err)
		_ = c.Error(err.(error)) // nolint: errcheck
		c.Abort()
		return
	}

	// 正常的 panic 处理逻辑
	httpRequest, _ := httputil.DumpRequest(c.Request, false)
	logx.Errorln("[Recovery from panic]",
		time.Now().UTC().Format(time.RFC3339),
		string(httpRequest),
		string(debug.Stack()),
		err,
	)
	c.AbortWithStatus(http.StatusInternalServerError)
}

func isBrokenPipeError(err interface{}) bool {
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			errMsg := strings.ToLower(se.Error())
			return strings.Contains(errMsg, "broken pipe") || strings.Contains(errMsg, "connection reset by peer")
		}
	}
	return false
}
