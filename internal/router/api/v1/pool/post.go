package pool

import (
	"errors"
	"net/http"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var req = new(types.Pool)
	if err := base.DecodeJson(r.Body, req); err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(err))
		return
	}
	if req.Size == 0 {
		base.SendJson(w, base.New().WithData(&types.Pool{
			Size:    worker.GetSize(),
			Running: worker.Running(),
			Waiting: worker.Waiting(),
		}))
		return
	}
	if (worker.Running() != 0 || worker.Waiting() != 0) && req.Size <= worker.GetSize() {
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(errors.New("there are still tasks running, scaling down is not allowed")))
		return
	}
	worker.SetSize(req.Size)
	base.SendJson(w, base.New().WithData(&types.Pool{
		Size:    worker.GetSize(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	}))
}
