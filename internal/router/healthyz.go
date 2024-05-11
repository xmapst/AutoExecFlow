package router

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/pkg/logx"
)

func healthyz(w http.ResponseWriter, r *http.Request) {
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
	data := &types.Healthyz{
		Server: r.Host,
		Client: r.RemoteAddr,
		State:  "Running",
	}
	if ws == nil {
		base.SendJson(w, base.New().WithData(data))
		return
	}
	// websocket 方式
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		err := ws.WriteJSON(base.New().WithData(data))
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
}
