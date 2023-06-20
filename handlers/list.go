package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/engine"
)

type ListRes struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	State   string `json:"state"`
	Code    int64  `json:"code"`
	Count   int64  `json:"count"`
	Message string `json:"msg"`
	Times   *Times `json:"times"`
}

type Times struct {
	Begin string `json:"begin,omitempty"`
	End   string `json:"end,omitempty"`
	TTL   string `json:"ttl,omitempty"`
}

// List
// @Summary List
// @description 列出当前所有任务ID
// @Tags Exec
// @Success 200 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router / [get]
func List(c *gin.Context) {
	render := Gin{Context: c}
	var tasksStates []*cache.ListData
	sortBy := c.DefaultQuery("sort", "begin")
	switch strings.ToLower(sortBy) {
	case "begin":
		tasksStates = cache.GetAllByBeginTime()
	case "end":
		tasksStates = cache.GetAllByEndTime()
	case "ttl":
		tasksStates = cache.GetAllByTTLTime()
	default:
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "sort参数错误",
		})
		return
	}
	if len(tasksStates) == 0 {
		render.SetJson(nil)
		return
	}
	var resSlice []ListRes
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	for _, v := range tasksStates {
		res := ListRes{
			ID:    v.ID,
			URL:   fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, v.ID),
			State: cache.StateCNMap[v.State],
			Count: v.Count,
			Times: &Times{
				Begin: timeStr(v.Times.Begin),
				End:   timeStr(v.Times.End),
			},
		}
		if v.Times.TTL > 0 {
			res.Times.TTL = v.Times.TTL.String()
		}
		if v.State == cache.SystemError {
			res.Code = 255
			res.Message = v.Message
		} else {
			tasksStepStates := cache.GetTaskAllStep(v.ID)
			var runningStateMsg, errorStateMsg []string
			var code int64
			for _, vv := range tasksStepStates {
				var state = fmt.Sprintf("步骤: %d, 名称: %s", vv.Step, vv.Name)
				if vv.Name == "" {
					state = fmt.Sprintf("步骤: %d", vv.Step)
				}
				if vv.State == cache.Running {
					runningStateMsg = append(runningStateMsg, state)
				}
				if vv.State == cache.Stop && vv.Code != 0 {
					errorStateMsg = append(errorStateMsg, state)
				}
				code += vv.Code
			}
			res.Code = code
			switch {
			case v.State == cache.Stop:
				res.Message = fmt.Sprintf("执行失败: [%s]", strings.Join(errorStateMsg, "; "))
				if res.Code == 0 {
					res.Message = "所有步骤执行成功"
				}
			case v.State == cache.Running:
				res.Message = fmt.Sprintf("当前正在执行: [%s]", strings.Join(runningStateMsg, "; "))
			default:
				res.Message = res.State
				if v.Message != "" {
					res.Message = v.Message
				}
			}
		}
		resSlice = append(resSlice, res)
	}
	render.SetJson(map[string]interface{}{
		"total":   len(resSlice),
		"tasks":   resSlice,
		"running": engine.QueueLength(),
	})
}
