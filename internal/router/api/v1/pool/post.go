package pool

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	var req types.Pool
	if err := c.ShouldBind(&req); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeFailed, err)
		return
	}
	if (worker.Running() != 0 || worker.Waiting() != 0) && req.Size <= worker.GetSize() {
		render.SetError(base.CodeFailed, errors.New("there are still tasks running, scaling down is not allowed"))
		return
	}
	worker.SetSize(req.Size)
	render.SetRes(&types.Pool{
		Size:    worker.GetSize(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	})
}
