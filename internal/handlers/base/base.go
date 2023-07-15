package base

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/handlers/types"
)

const (
	CodeErrParam = iota + 1000
	CodeRunning
	CodeExecErr
	CodeErrNoData
	CodePending
	CodeSuccess = 0
	CodeErrApp  = 500
)

var MsgFlags = map[int64]string{
	CodeErrParam:  "parameter error",
	CodeRunning:   "running",
	CodeExecErr:   "exec error",
	CodeErrNoData: "no data",
	CodePending:   "pending",
	CodeSuccess:   "success",
	CodeErrApp:    "internal error",
}

// GetMsg get error information based on Code
func GetMsg(code int64) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[CodeErrApp]
}

type Gin struct {
	*gin.Context
}

func NewRes(data interface{}, err error, code int64) *types.BaseRes {
	if code == 200 {
		code = 0
	}
	codeMsg := GetMsg(code)
	return &types.BaseRes{
		Data: data,
		Code: code,
		Message: func() string {
			result := NewInfo(err)
			if codeMsg != "" && result != "" {
				if !strings.HasPrefix(result, codeMsg) {
					result += "; " + codeMsg
				}
			} else if codeMsg != "" {
				result = codeMsg
			}
			return strings.TrimSpace(result)
		}(),
	}
}
func NewInfo(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// SetRes Response res
func (g *Gin) SetRes(res interface{}, err error, code int64) {
	g.JSON(http.StatusOK, NewRes(res, err, code))
}

// SetJson Set Json
func (g *Gin) SetJson(res interface{}) {
	g.SetRes(res, nil, CodeSuccess)
}

// SetError Check Error
func (g *Gin) SetError(code int64, err error) {
	g.SetRes(nil, err, code)
	g.Abort()
}
