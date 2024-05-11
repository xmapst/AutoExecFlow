package pool

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Detail(w http.ResponseWriter, r *http.Request) {
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(r) {
		var err error
		ws, err = base.Upgrade(w, r)
		if err != nil {
			logx.Errorln(err)
			base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
			return
		}
	}
	if ws == nil {
		base.SendJson(w, base.New().WithCode(base.CodeSuccess).WithData(&types.Pool{
			Size:    worker.GetSize(),
			Total:   worker.GetTotal(),
			Running: worker.Running(),
			Waiting: worker.Waiting(),
		}))
		return
	}
	// websocket 方式
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		res := &types.Pool{
			Size:    worker.GetSize(),
			Total:   worker.GetTotal(),
			Running: worker.Running(),
			Waiting: worker.Waiting(),
		}
		err := ws.WriteJSON(base.New().WithData(res))
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
}
