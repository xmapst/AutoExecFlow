package base

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xmapst/osreapi/internal/utils"
)

func DefaultQuery(r *http.Request, key, defaultValue string) string {
	value := Query(r, key)
	if value == "" {
		value = defaultValue
	}
	return value
}

func Query(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func DecodeJson(r io.Reader, v interface{}) error {
	defer io.Copy(io.Discard, r) //nolint:errcheck
	return json.NewDecoder(r).Decode(v)
}

func SendJson(w http.ResponseWriter, v interface{}) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes()) //nolint:errcheck
}

type Base interface {
	WithCode(code int) Base
	WithData(data interface{}) Base
	WithError(err error) Base
}

type res struct {
	Code      int         `json:"code" yaml:"Code" toml:"code" example:"255"`
	Message   Message     `json:"msg" yaml:"Message" toml:"message" example:"message"`
	Timestamp int64       `json:"timestamp" yaml:"Timestamp" toml:"timestamp"`
	Data      interface{} `json:"data" yaml:"Data" toml:"data"`
}

type Message []string

func (msg *Message) String() string {
	return strings.Join(utils.RemoveDuplicate(*msg), "; ")
}

func (msg *Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(msg.String())
}

func (msg *Message) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	for _, v := range strings.Split(str, "; ") {
		*msg = append(*msg, v)
	}
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
func (r *res) getMsg() string {
	msg, ok := msgFlags[r.Code]
	if !ok {
		msg = msgFlags[CodeNoData]
	}
	return msg
}

func New() Base {
	r := &res{
		Timestamp: time.Now().UnixNano(),
	}
	r.Message = append(r.Message, r.getMsg())
	return r
}

func (r *res) isExistCode(s string) bool {
	for _, v := range msgFlags {
		if v == s {
			return true
		}
	}
	return false
}

func (r *res) WithCode(code int) Base {
	if code == http.StatusOK {
		code = 0
	}
	// clear old code message
	var msg Message
	for _, v := range r.Message {
		if r.isExistCode(v) {
			continue
		}
		msg = append(msg, v)
	}
	r.Code = code
	r.Message = append(msg, r.getMsg())
	r.Timestamp = time.Now().UnixNano()
	return r
}

func (r *res) WithData(data interface{}) Base {
	r.Data = data
	r.Timestamp = time.Now().UnixNano()
	return r
}

func (r *res) WithError(err error) Base {
	if err == nil {
		return r
	}
	// clear old code message
	var msg Message
	for _, v := range r.Message {
		if r.isExistCode(v) {
			continue
		}
		msg = append(msg, v)
	}
	r.Message = append(msg, strings.TrimSpace(err.Error()))
	r.Timestamp = time.Now().UnixNano()
	return r
}