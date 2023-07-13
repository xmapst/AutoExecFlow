package task

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/utils"
)

type ListRes struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	State   int    `json:"state"`
	Code    int64  `json:"code"`
	Count   int64  `json:"count"`
	Message string `json:"msg"`
	Times   *Times `json:"times"`
}

type Times struct {
	ST string `json:"st,omitempty"` // 开始时间
	ET string `json:"et,omitempty"` // 结束时间
	RT string `json:"rt,omitempty"` // 剩余时间
}

// List
// @Summary task detail
// @description get all task
// @Tags Task
// @Accept json
// @Produce json
// @Param sort query string false "sort param" Enums(st,et,rt) default(st)
// @Success 200 {object} base.Result
// @Failure 500 {object} base.Result
// @Router /api/v1/task [get]
func List(c *gin.Context) {
	render := base.Gin{Context: c}
	var tasksStates []*cache.ListData
	sortBy := c.DefaultQuery("sort", "st")
	switch strings.ToLower(sortBy) {
	case "st":
		tasksStates = cache.GetAllByBeginTime()
	case "et":
		tasksStates = cache.GetAllByEndTime()
	case "rt":
		tasksStates = cache.GetAllByTTLTime()
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "sort parameter error",
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
			URL:   fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), v.ID),
			State: v.State,
			Count: v.Count,
			Times: &Times{
				ST: utils.TimeStr(v.Times.ST),
				ET: utils.TimeStr(v.Times.ET),
			},
		}
		if v.Times.RT > 0 {
			res.Times.RT = v.Times.RT.String()
		}
		if v.State == cache.SystemError {
			res.Code = 255
			res.Message = v.Message
		} else {
			tasksStepStates := cache.GetTaskStepStates(v.ID)
			var runningStateMsg, errorStateMsg []string
			var code int64
			for _, vv := range tasksStepStates {
				var state = fmt.Sprintf("Step: %d, Name: %s", vv.Step, vv.Name)
				if vv.Name == "" {
					state = fmt.Sprintf("Step: %d", vv.Step)
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
			res.Message = cache.StateMap[v.State]
			switch {
			case v.State == cache.Stop:
				res.Message = fmt.Sprintf("execution failed: [%s]", strings.Join(errorStateMsg, "; "))
				if res.Code == 0 {
					res.Message = "all steps executed successfully"
				}
			case v.State == cache.Running:
				res.Message = fmt.Sprintf("currently executing: [%s]", strings.Join(runningStateMsg, "; "))
			default:
				if v.Message != "" {
					res.Message = v.Message
				}
			}
		}
		resSlice = append(resSlice, res)
	}
	render.SetJson(gin.H{
		"total":     len(resSlice),
		"tasks":     resSlice,
		"timestamp": time.Now().UnixNano(),
	})
}
