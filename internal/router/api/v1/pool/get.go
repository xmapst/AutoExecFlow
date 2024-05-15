package pool

import (
	"net/http"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
)

func Detail(w http.ResponseWriter, r *http.Request) {
	base.SendJson(w, base.New().WithCode(base.CodeSuccess).WithData(&types.Pool{
		Size:    worker.GetSize(),
		Total:   worker.GetTotal(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	}))
}
