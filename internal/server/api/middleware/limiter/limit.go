package limiter

import (
	"net"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/types"
)

// key 是 ip+mode+path  value ==>bucket
var cache = new(sync.Map)

func New(rate int64, modes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rate <= 0 {
			return
		}
		var exist bool
		for _, mode := range modes {
			exist = c.Request.Method == mode
			if exist {
				break
			}
		}
		if exist {
			ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
			key := c.Request.Method + ip + c.Request.RequestURI
			var limiter *Bucket
			if v, ok := cache.Load(key); ok {
				limiter = v.(*Bucket)
			} else {
				limiter = newBucket(rate)
				cache.Store(key, limiter)
			}

			if !limiter.IsAccept() {
				base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("this ip too many requests")))
			}
		}
	}
}
