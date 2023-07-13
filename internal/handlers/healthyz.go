package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func healthyz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"server":    c.Request.Host,
		"client":    c.ClientIP(),
		"state":     "Running",
		"timestamp": time.Now().UnixNano(),
	})
}
