package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/utils"
	"net/http"
	"strings"
	"time"
)

const (
	xTaskID    = "X-Task-ID"
	xTaskState = "X-Task-STATE"
)

type Gin struct {
	*gin.Context
}

type JSONResult struct {
	Code    int         `json:"code" description:"返回码" example:"0000"`
	Message string      `json:"message,omitempty" description:"消息" example:"消息"`
	Data    interface{} `json:"data,omitempty" description:"数据"`
}

func NewRes(data interface{}, err error, code int) *JSONResult {
	if code == 200 {
		code = 0
	}
	codeMsg := utils.GetMsg(code)
	return &JSONResult{
		Data: data,
		Code: code,
		Message: func() string {
			result := NewInfo(err)
			if codeMsg != "" && result != "" {
				result += ", " + codeMsg
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
func (g *Gin) SetRes(res interface{}, err error, code int) {
	g.JSON(http.StatusOK, NewRes(res, err, code))
}

// SetJson Set Json
func (g *Gin) SetJson(res interface{}) {
	g.SetRes(res, nil, utils.CodeSuccess)
}

// SetError Check Error
func (g *Gin) SetError(code int, err error) {
	g.SetRes(nil, err, code)
	g.Abort()
}

func timeStr(nsec int64) string {
	if nsec == 0 {
		return ""
	}
	return time.Unix(0, nsec).Format(time.RFC3339)
}
