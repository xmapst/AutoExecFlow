package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/gorilla/websocket"

	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/pkg/pty"
	"github.com/xmapst/AutoExecFlow/types"
)

type PtyService struct {
	ws  *websocket.Conn
	pty pty.Terminal
}

func Pty(ws *websocket.Conn) (*PtyService, error) {
	terminal, err := pty.New("")
	if err != nil {
		logx.Errorln(err)
		return nil, err
	}
	return &PtyService{
		pty: terminal,
		ws:  ws,
	}, nil
}

func (p *PtyService) Run() {
	go func() {
		_, _ = io.Copy(p.pty, p)
	}()
	go func() {
		_, _ = p.pty.Wait(context.Background())
		_ = p.ws.Close()
	}()
	_, _ = io.Copy(p, p.pty)
	_ = p.ws.Close()
}

func (p *PtyService) Read(b []byte) (n int, err error) {
	messageType, data, err := p.ws.ReadMessage()
	if err != nil {
		return
	}

	dataBuffer := bytes.Trim(data, "\x00")
	_, ok := types.WebsocketMessageType[messageType]
	if !ok {
		return
	}
	if messageType == websocket.BinaryMessage {
		if dataBuffer[0] == 1 {
			ttySize := &types.TTYSize{}
			resizeMessage := bytes.Trim(dataBuffer[1:], " \n\r\t\x00\x01")
			if err = json.Unmarshal(resizeMessage, ttySize); err != nil {
				return
			}
			_ = p.pty.Resize(int16(ttySize.Rows), int16(ttySize.Cols))
			return
		}
	}

	return copy(b, dataBuffer), nil
}

func (p *PtyService) Write(b []byte) (n int, err error) {
	n = len(b)
	err = p.ws.WriteMessage(websocket.BinaryMessage, b)
	return
}
