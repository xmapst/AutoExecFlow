package event

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/pkg/dag"
)

func Stream(c *gin.Context) {
	event, id, err := dag.SubscribeEvent()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer dag.UnSubscribeEvent(id)
	ticker := time.NewTicker(30 * time.Second) // 每30秒发送心跳
	defer ticker.Stop()
	var once sync.Once
	c.Stream(func(w io.Writer) bool {
		once.Do(func() {
			// send event heartbeat at the beginning
			c.SSEvent("heartbeat", "keepalive")
		})
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
