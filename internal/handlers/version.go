package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	info "github.com/xmapst/osreapi"
)

func version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": info.Version,
		"go": gin.H{
			"version": runtime.Version(),
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		},
		"git": gin.H{
			"url":    info.GitUrl,
			"branch": info.GitBranch,
			"commit": info.GitCommit,
		},
		"user": gin.H{
			"name":  info.UserName,
			"email": info.UserEmail,
		},
		"build_time": info.BuildTime,
		"timestamp":  time.Now().UnixNano(),
	})
}
