package event

import (
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/types"
)

func Stream(c *gin.Context) {
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	var event = make(chan string, 65534)
	err := service.Event().Subscribe(ctx, event)
	if err != nil {
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
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
