package sys

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	_ "github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/pkg/pty"
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

// PtyWs
// @Summary		Terminal
// @description	websocket terminal
// @Tags		System
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.Base[any]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/pty [get]
func PtyWs(c *gin.Context) {
	ws, err := base.Upgrade(c.Writer, c.Request)
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
