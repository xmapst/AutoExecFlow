package sys

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/pkg/logx"
	"github.com/xmapst/osreapi/pkg/pty"
)

var WebsocketMessageType = map[int]string{
	websocket.BinaryMessage: "binary",
	websocket.TextMessage:   "text",
	websocket.CloseMessage:  "close",
	websocket.PingMessage:   "ping",
	websocket.PongMessage:   "pong",
}

type TTYSize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	X    uint16 `json:"x"`
	Y    uint16 `json:"y"`
}

func PtyWs(w http.ResponseWriter, r *http.Request) {
	ws, err := base.Upgrade(w, r)
	if err != nil {
		logx.Errorln(err)
		return
	}
	defer ws.Close()

	_pty, err := pty.New("")
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), time.Now().Add(time.Second))
		return
	}
	_tun := &tun{
		ws:  ws,
		pty: _pty,
	}

	go func() {
		_, _ = io.Copy(_pty, _tun)
	}()
	go func() {
		_, _ = _pty.Wait(context.Background())
		_ = ws.Close()
	}()
	_, _ = io.Copy(_tun, _pty)
	_ = ws.Close()
}

type tun struct {
	ws  *websocket.Conn
	pty pty.Terminal
}

func (t *tun) Read(p []byte) (n int, err error) {
	messageType, data, err := t.ws.ReadMessage()
	if err != nil {
		return
	}

	dataBuffer := bytes.Trim(data, "\x00")
	_, ok := WebsocketMessageType[messageType]
	if !ok {
		return
	}
	if messageType == websocket.BinaryMessage {
		if dataBuffer[0] == 1 {
			ttySize := &TTYSize{}
			resizeMessage := bytes.Trim(dataBuffer[1:], " \n\r\t\x00\x01")
			if err = json.Unmarshal(resizeMessage, ttySize); err != nil {
				return
			}
			_ = t.pty.Resize(int16(ttySize.Rows), int16(ttySize.Cols))
			return
		}
	}

	return copy(p, dataBuffer), nil
}

func (t *tun) Write(p []byte) (n int, err error) {
	n = len(p)
	err = t.ws.WriteMessage(websocket.BinaryMessage, p)
	return
}
