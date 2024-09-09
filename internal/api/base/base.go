package base

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/xmapst/AutoExecFlow/types"
)

type IBase[T any] interface {
	WithCode(code types.Code) IBase[T]
	WithError(err error) IBase[T]
	WithData(data T) IBase[T]
}

func Send(g *gin.Context, v interface{}) {
	switch g.NegotiateFormat(binding.MIMEJSON, binding.MIMEYAML, binding.MIMEYAML2) {
	case binding.MIMEJSON:
		g.JSON(http.StatusOK, v)
	case binding.MIMEYAML, binding.MIMEYAML2:
		g.YAML(http.StatusOK, v)
	default:
		g.JSON(http.StatusOK, v)
	}
}

type Base[T any] struct {
	types.Base[T]
}

// getMsg get error information based on Code
func (r *Base[T]) getMsg() string {
	msg, ok := types.CodeMap[r.Code]
	if !ok {
		msg = types.CodeMap[types.CodeNoData]
	}
	return msg
}

func WithCode[T any](code types.Code) IBase[T] {
	r := &Base[T]{
		Base: types.Base[T]{
			Timestamp: time.Now().UnixNano(),
		},
	}
	return r.WithCode(code)
}

func WithData[T any](data T) IBase[T] {
	r := &Base[T]{
		Base: types.Base[T]{
			Code:      types.CodeSuccess,
			Timestamp: time.Now().UnixNano(),
		},
	}
	return r.WithData(data)
}

func WithError[T any](err error) IBase[T] {
	r := &Base[T]{
		Base: types.Base[T]{
			Code:      types.CodeFailed,
			Timestamp: time.Now().UnixNano(),
		},
	}
	return r.WithError(err)
}

// Removes existing code messages from the Message slice.
func (r *Base[T]) removeMsgCodes() types.Message {
	var msg types.Message
	for _, v := range r.Message {
		if !r.isCodeMsg(v) {
			msg = append(msg, v)
		}
	}
	return msg
}

// Checks if the message exists in msgFlags.
func (r *Base[T]) isCodeMsg(message string) bool {
	for _, v := range types.CodeMap {
		if v == message {
			return true
		}
	}
	return false
}

func (r *Base[T]) WithCode(code types.Code) IBase[T] {
	if code == http.StatusOK {
		code = types.CodeSuccess
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
		r.Message = types.Message{r.getMsg()}
	}
	r.Data = data
	r.Timestamp = time.Now().UnixNano()
	return r
}
