package event

import (
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/service"
)

func Stream(c *gin.Context) {
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	var event = make(chan string, 65534)
	service.Event().Subscribe(ctx, event)
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送心跳
	defer ticker.Stop()
	c.Stream(func(w io.Writer) bool {
		select {
		case e, ok := <-event:
			if !ok {
				return false
			}
			c.SSEvent("message", e)
			return true
		case <-ticker.C:
			c.SSEvent("heartbeat", "keepalive")
			return true
		case <-c.Writer.CloseNotify():
			return false
		}
	})
}
