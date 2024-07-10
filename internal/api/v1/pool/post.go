package pool

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

// Post
// @Summary		Setting
// @description	Configuring the Task Pool Size
// @Tags		Pool
// @Accept		application/json
// @Accept		application/yaml
// @Accept		multipart/form-data
// @Produce		application/json
// @Produce		application/yaml
// @Param		setting body types.Pool true "pool setting"
// @Success		200 {object} types.Base[types.Pool]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/pool [post]
func Post(c *gin.Context) {
	var req = new(types.Pool)
	if err := c.ShouldBind(&req); err != nil {
		logx.Errorln(err)
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	if req.Size <= 0 {
		base.Send(c, types.WithData(&types.Pool{
			Size:    worker.GetSize(),
			Total:   worker.GetTotal(),
			Running: worker.Running(),
			Waiting: worker.Waiting(),
		}))
		return
	}
	if (worker.Running() != 0 || worker.Waiting() != 0) && req.Size <= worker.GetSize() {
		base.Send(c, types.WithCode[any](types.CodeFailed).WithError(errors.New("there are still tasks running, scaling down is not allowed")))
		return
	}
	worker.SetSize(req.Size)
	base.Send(c, types.WithData(&types.Pool{
		Size:    worker.GetSize(),
		Total:   worker.GetTotal(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	}),
	)
}
