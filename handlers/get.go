package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/utils"
	"net/http"
	"strings"
)

type ResStatus struct {
	Step      int64    `json:"step"`
	Name      string   `json:"name,omitempty"`
	State     string   `json:"state"`
	Code      int64    `json:"code"`
	Message   []string `json:"message"`
	DependsOn []string `json:"depends_on,omitempty"`
	Times     *Times   `json:"times,omitempty"`
}

// Get
// @Summary 查询
// @description 查询执行情况
// @Tags Exec
// @Param id path string true "id"
// @Success 200 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router /{id} [get]
func Get(c *gin.Context) {
	render := Gin{Context: c}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": "缺少id参数",
		})
		return
	}
	taskState, found := cache.GetTaskState(id)
	if !found {
		render.SetError(utils.CodeErrNoData, errors.New("id不存在"))
		return
	}
	state := cache.StateENMap[taskState.State]
	c.Request.Header.Set(xTaskState, state)
	c.Writer.Header().Set(xTaskState, state)
	c.Set(xTaskState, state)
	tasksStepStates := cache.GetAllTaskStepState(id)
	var res []ResStatus
	for _, v := range tasksStepStates {
		var output []string
		outputs := cache.GetAllTaskStepOutput(id, v.Step)
		for _, o := range outputs {
			output = append(output, o.Content)
		}
		_res := ResStatus{
			Step:      v.Step,
			Name:      v.Name,
			State:     cache.StateCNMap[v.State],
			Code:      v.Code,
			DependsOn: v.DependsOn,
			Times: &Times{
				Begin: timeStr(v.Times.Begin),
				End:   timeStr(v.Times.End),
				TTL:   v.Times.TTL.String(),
			},
		}
		if v.Message != "" {
			_res.Message = append(_res.Message, v.Message)
		}
		if output != nil {
			_res.Message = append(_res.Message, output...)
		}
		res = append(res, _res)
	}

	switch taskState.State {
	// 运行结束
	case cache.Stop:
		var code int64
		var message []string
		for _, v := range tasksStepStates {
			code += v.Code
			if v.Code != 0 {
				var msg = fmt.Sprintf("步骤: %d, 退出码: %d", v.Step, v.Code)
				if v.Name != "" {
					msg = fmt.Sprintf("步骤: %d, 名称: %s, 退出码: %d", v.Step, v.Name, v.Code)
				}
				if taskState.VMInstanceID != "" {
					msg += fmt.Sprintf(", 实例ID: %s, %s", taskState.VMInstanceID, message)
				}
				if taskState.HardWareID != "" {
					msg += fmt.Sprintf(", 硬件ID: %s, %s", taskState.HardWareID, message)
				}
				message = append(message, msg)
			}
		}
		if code != 0 {
			render.SetRes(res, fmt.Errorf("执行失败: [%s]", strings.Join(message, "; ")), utils.CodeExecErr)
			return
		}
		render.SetJson(res)
	// 运行中, 排队中
	case cache.Running:
		var msgSlice []string
		for _, v := range tasksStepStates {
			if v.State != cache.Running {
				continue
			}
			var msg = fmt.Sprintf("步骤: %d, 名称: %s", v.Step, v.Name)
			if v.Name == "" {
				msg = fmt.Sprintf("步骤: %d", v.Step)
			}
			msgSlice = append(msgSlice, msg)
		}
		render.SetRes(res, fmt.Errorf("执行中: [%s]", strings.Join(msgSlice, "; ")), utils.CodeRunning)
	case cache.Pending:
		render.SetError(utils.CodeRunning, nil)
	case cache.SystemError:
		render.SetRes(res, fmt.Errorf(taskState.Message), utils.CodeExecErr)
	default:
		render.SetError(utils.CodeErrNoData, errors.New("id不存在"))
	}
}
