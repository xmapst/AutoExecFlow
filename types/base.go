package types

import (
	"encoding/json"
	"strings"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"

	"github.com/xmapst/AutoExecFlow/internal/utils"
)

type Base[T any] struct {
	Code      Code    `json:"code" yaml:"Code" example:"255"`
	Message   Message `json:"msg" yaml:"Message" example:"message" swaggertype:"string"`
	Timestamp int64   `json:"timestamp" yaml:"Timestamp"`
	Data      T       `json:"data" yaml:"Data"`
}

type Message []string

func (msg Message) String() string {
	return strings.Join(utils.RemoveDuplicate(msg), "; ")
}

func (msg Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(msg.String())
}

func (msg Message) MarshalYAML() (interface{}, error) {
	data, err := yaml.Marshal(msg.String())
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (msg *Message) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*msg = strings.Split(s, "; ")
	return nil
}

func (msg *Message) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*msg = strings.Split(s, "; ")
	return nil
}

type Code int

const (
	CodeSuccess Code = 0
	CodeRunning Code = iota + 1000
	CodeFailed
	CodeNoData
	CodePending
	CodePaused
)

var CodeMap = map[Code]string{
	CodeSuccess: "success",
	CodeRunning: "running",
	CodeFailed:  "failed",
	CodeNoData:  "no data",
	CodePending: "pending",
	CodePaused:  "paused",
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