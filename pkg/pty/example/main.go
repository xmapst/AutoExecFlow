package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/pkg/logx"
	"github.com/xmapst/osreapi/pkg/pty"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	go func() {
		<-sigs
		os.Exit(0)
	}()

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Group(func(r chi.Router) {
		fs := http.StripPrefix("/", http.FileServer(http.Dir("ui")))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		})
	})
	r.Get("/ws", tty)
	_ = http.ListenAndServe(":8081", r)
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

func tty(w http.ResponseWriter, r *http.Request) {
	ws, err := upGrader.Upgrade(w, r, nil)
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
	tun := &Tun{
		ws:  ws,
		pty: _pty,
	}

	go func() {
		_, _ = io.Copy(_pty, tun)
	}()
	go func() {
		_, _ = _pty.Wait(context.Background())
		_ = ws.Close()
	}()
	_, _ = io.Copy(tun, _pty)

}

type Tun struct {
	ws  *websocket.Conn
	pty pty.Terminal
}

func (t *Tun) Read(p []byte) (n int, err error) {
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

func (t *Tun) Write(p []byte) (n int, err error) {
	n = len(p)
	err = t.ws.WriteMessage(websocket.BinaryMessage, p)
	return
}
