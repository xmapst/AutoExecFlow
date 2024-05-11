package pool

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
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
			render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
			return
		}
	}
	if ws == nil {
		render.JSON(w, r, types.New().WithCode(types.CodeSuccess).WithData(&types.Pool{
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
		err := ws.WriteJSON(types.New().WithData(res))
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
}
