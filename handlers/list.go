package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/engine"
	"net/http"
	"strings"
)

type ListRes struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Code    int    `json:"code"`
	Count   int    `json:"count"`
	Message string `json:"message"`
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
	for _, taskState := range tasksStates {
		res := ListRes{
			ID:    taskState.ID,
			State: cache.StateCNMap[taskState.State],
			Count: taskState.Count,
			Times: &Times{
				Begin: timeStr(taskState.Times.Begin),
				End:   timeStr(taskState.Times.End),
			},
		}
		if taskState.Times.TTL > 0 {
			res.Times.TTL = taskState.Times.TTL.String()
		}
		if taskState.State == cache.SystemError {
			res.Code = 255
			res.Message = "系统错误"
		} else {
			tasksStepStates := cache.GetAllTaskStepState(taskState.ID)
			var runningStateMsg, errorStateMsg []string
			var code int
			for _, v := range tasksStepStates {
				var state = fmt.Sprintf("步骤: %d, 名称: %s", v.Step, v.Name)
				if v.Name == "" {
					state = fmt.Sprintf("步骤: %d", v.Step)
				}
				if v.State == cache.Running {
					runningStateMsg = append(runningStateMsg, state)
				}
				if v.State == cache.Stop && v.Code != 0 {
					errorStateMsg = append(errorStateMsg, state)
				}
				code += v.Code
			}
			res.Code = code
			switch {
			case taskState.State == cache.Stop:
				res.Message = fmt.Sprintf("执行失败: [%s]", strings.Join(errorStateMsg, "; "))
				if res.Code == 0 {
					res.Message = "所有步骤执行成功"
				}
			case taskState.State == cache.Running:
				res.Message = fmt.Sprintf("当前正在执行: [%s]", strings.Join(runningStateMsg, "; "))
			default:
				res.Message = res.State
			}
		}
		resSlice = append(resSlice, res)
	}
	render.SetJson(map[string]interface{}{
		"total":   len(resSlice),
		"tasks":   resSlice,
		"running": engine.Pool.QueueLength(),
	})
}