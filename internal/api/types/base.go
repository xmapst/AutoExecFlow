package types

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/xmapst/AutoExecFlow/internal/utils"
)

type IBase[T any] interface {
	WithCode(code int) IBase[T]
	WithError(err error) IBase[T]
	WithData(data T) IBase[T]
}

type Base[T any] struct {
	Code      int     `json:"code" yaml:"Code" example:"255"`
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

const (
	CodeSuccess = 0
	CodeRunning = iota + 1000
	CodeFailed
	CodeNoData
	CodePending
	CodePaused
)

var msgFlags = map[int]string{
	CodeSuccess: "success",
	CodeRunning: "running",
	CodeFailed:  "failed",
	CodeNoData:  "no data",
	CodePending: "pending",
	CodePaused:  "paused",
}

// getMsg get error information based on Code
func (r *Base[T]) getMsg() string {
	msg, ok := msgFlags[r.Code]
	if !ok {
		msg = msgFlags[CodeNoData]
	}
	return msg
}

func WithCode[T any](code int) IBase[T] {
	r := &Base[T]{
		Timestamp: time.Now().UnixNano(),
	}
	return r.WithCode(code)
}

func WithData[T any](data T) IBase[T] {
	r := &Base[T]{
		Timestamp: time.Now().UnixNano(),
	}
	return r.WithData(data)
}

func WithError[T any](err error) IBase[T] {
	r := &Base[T]{
		Timestamp: time.Now().UnixNano(),
	}
	return r.WithError(err)
}

func (r *Base[T]) isExistCode(s string) bool {
	for _, v := range msgFlags {
		if v == s {
			return true
		}
	}
	return false
}

// Removes existing code messages from the Message slice.
func (r *Base[T]) removeMsgCodes() Message {
	var msg Message
	for _, v := range r.Message {
		if !r.isCodeMsg(v) {
			msg = append(msg, v)
		}
	}
	return msg
}

// Checks if the message exists in msgFlags.
func (r *Base[T]) isCodeMsg(message string) bool {
	for _, v := range msgFlags {
		if v == message {
			return true
		}
	}
	return false
}

func (r *Base[T]) WithCode(code int) IBase[T] {
	if code == http.StatusOK {
		code = CodeSuccess
	}
	r.Code = code
	r.Message = append(r.removeMsgCodes(), r.getMsg())
	r.Timestamp = time.Now().UnixNano()
	return r
}

func (r *Base[T]) WithError(err error) IBase[T] {
	if err == nil {
		return r
	}
	r.Message = append(r.removeMsgCodes(), strings.TrimSpace(err.Error()))
	r.Timestamp = time.Now().UnixNano()
	return r
}

func (r *Base[T]) WithData(data T) IBase[T] {
	if r.Message == nil {
		r.Message = Message{r.getMsg()}
	}
	r.Data = data
	r.Timestamp = time.Now().UnixNano()
	return r
}
