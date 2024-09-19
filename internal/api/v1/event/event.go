package event

import (
	"io"
	"net/http"

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
	c.Stream(func(w io.Writer) bool {
		e, ok := <-event
		if !ok {
			return false
		}
		c.SSEvent("message", e)
		return true
	})
}
