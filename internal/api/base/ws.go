package base

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024 * 1024 * 10,
	HandshakeTimeout: 3 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upGrader.Upgrade(w, r, nil)
}
