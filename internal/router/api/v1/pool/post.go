package pool

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var req = new(types.Pool)
	if err := render.DecodeJSON(r.Body, req); err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(err))
		return
	}
	if req.Size == 0 {
		render.JSON(w, r, types.New().WithData(&types.Pool{
			Size:    worker.GetSize(),
			Running: worker.Running(),
			Waiting: worker.Waiting(),
		}))
		return
	}
	if (worker.Running() != 0 || worker.Waiting() != 0) && req.Size <= worker.GetSize() {
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(errors.New("there are still tasks running, scaling down is not allowed")))
		return
	}
	worker.SetSize(req.Size)
	render.JSON(w, r, types.New().WithData(&types.Pool{
		Size:    worker.GetSize(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	}))
}
