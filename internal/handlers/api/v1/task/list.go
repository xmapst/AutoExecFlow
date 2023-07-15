package task

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/exec"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/handlers/types"
	"github.com/xmapst/osreapi/internal/utils"
)

// List
// @Summary task detail
// @description get all task
// @Tags Task
// @Accept json
// @Produce json
// @Param sort query string false "sort param" Enums(st,et,rt) default(st)
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
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
	var resSlice []types.TaskListRes
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	for _, v := range tasksStates {
		res := types.TaskListRes{
			ID:    v.ID,
			URL:   fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), v.ID),
			State: v.State,
			Count: v.Count,
			Times: &types.Times{
				ST: utils.TimeStr(v.Times.ST),
				ET: utils.TimeStr(v.Times.ET),
			},
		}
		if v.Times.RT > 0 {
			res.Times.RT = v.Times.RT.String()
		}
		if v.State == exec.SystemErr {
			res.Code = exec.SystemErr
			res.Message = v.Message
		} else {
			tasksStepStates := cache.GetTaskStepStates(v.ID)
			var runningStateMsg, errorStateMsg []string
			var code int64
			for _, vv := range tasksStepStates {
				var state = fmt.Sprintf("Step: %d, Name: %s", vv.ID, vv.Name)
				if vv.Name == "" {
					state = fmt.Sprintf("Step: %d", vv.ID)
				}
				if vv.State == exec.Running {
					runningStateMsg = append(runningStateMsg, state)
				}
				if vv.State == exec.Stop && vv.Code != 0 {
					errorStateMsg = append(errorStateMsg, state)
				}
				code += vv.Code
			}
			res.Code = code
			res.Message = exec.StateMap[v.State]
			switch {
			case v.State == exec.Stop:
				res.Message = fmt.Sprintf("execution failed: [%s]", strings.Join(errorStateMsg, "; "))
				if res.Code == 0 {
					res.Message = "all steps executed successfully"
				}
			case v.State == exec.Running:
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
