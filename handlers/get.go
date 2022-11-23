package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ResStatus struct {
	Step      int64    `json:"step"`
	URL       string   `json:"url,omitempty"`
	Name      string   `json:"name,omitempty"`
	State     string   `json:"state"`
	Code      int64    `json:"code"`
	Message   string   `json:"message"`
	DependsOn []string `json:"depends_on,omitempty"`
	Times     *Times   `json:"times,omitempty"`
}

// GetTask
// @Summary 查询任务
// @description 查询任务执行情况
// @Tags Exec
// @Param id path string true "id"
// @Success 200 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router /{id} [get]
func GetTask(c *gin.Context) {
	render := Gin{Context: c}
	id := c.Param("id")
	if id == "" {
		render.SetError(utils.CodeErrNoData, errors.New("任务不存在"))
		return
	}
	taskState, found := cache.GetTask(id)
	if !found {
		render.SetError(utils.CodeErrNoData, errors.New("任务不存在"))
		return
	}
	state := cache.StateENMap[taskState.State]
	c.Request.Header.Set(xTaskState, state)
	c.Writer.Header().Set(xTaskState, state)
	c.Set(xTaskState, state)
	tasksStepStates := cache.GetTaskAllStep(id)
	var res []ResStatus
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	var code int64
	var stopMsg, runMsg []string
	for _, v := range tasksStepStates {
		//var output []string
		//outputs := cache.GetAllTaskStepOutput(id, v.Step)
		//for _, o := range outputs {
		//	output = append(output, o.Content)
		//}
		code += v.Code
		msg := v.Message
		if v.Code != 0 {
			msg = fmt.Sprintf("步骤: %d, 退出码: %d", v.Step, v.Code)
			if v.Name != "" {
				msg = fmt.Sprintf("步骤: %d, 名称: %s, 退出码: %d", v.Step, v.Name, v.Code)
			}
			if taskState.VMInstanceID != "" {
				msg += fmt.Sprintf(", 实例ID: %s", taskState.VMInstanceID)
			}
			if taskState.HardWareID != "" {
				msg += fmt.Sprintf(", 硬件ID: %s", taskState.HardWareID)
			}
			stopMsg = append(stopMsg, msg)
		}
		if v.State == cache.Running {
			msg = fmt.Sprintf("步骤: %d, 名称: %s", v.Step, v.Name)
			if v.Name == "" {
				msg = fmt.Sprintf("步骤: %d", v.Step)
			}
			runMsg = append(runMsg, msg)
		}
		_res := ResStatus{
			Step:      v.Step,
			URL:       fmt.Sprintf("%s://%s/%s/%d/console", scheme, c.Request.Host, id, v.Step),
			Name:      v.Name,
			State:     cache.StateCNMap[v.State],
			Code:      v.Code,
			DependsOn: v.DependsOn,
			Message:   msg,
			Times: &Times{
				Begin: timeStr(v.Times.Begin),
				End:   timeStr(v.Times.End),
				TTL:   v.Times.TTL.String(),
			},
		}
		res = append(res, _res)
	}

	switch taskState.State {
	// 运行结束
	case cache.Stop:
		if code != 0 {
			render.SetRes(res, fmt.Errorf("执行失败: [%s]", strings.Join(stopMsg, "; ")), utils.CodeExecErr)
			return
		}
		render.SetJson(res)
	// 运行中, 排队中
	case cache.Running:
		render.SetRes(res, fmt.Errorf("执行中: [%s]", strings.Join(runMsg, "; ")), utils.CodeRunning)
	case cache.Pending:
		render.SetError(utils.CodeRunning, nil)
	case cache.SystemError:
		render.SetRes(res, fmt.Errorf(taskState.Message), utils.CodeExecErr)
	default:
		render.SetError(utils.CodeErrNoData, errors.New("任务不存在"))
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// GetStep
// @Summary 查询步骤
// @description 查询步骤执行情况
// @Tags Exec
// @Param id path string true "id"
// @Param step path string true "step"
// @Success 200 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router /{id}/{step}/console [get]
func GetStep(c *gin.Context) {
	render := Gin{Context: c}
	id := c.Param("id")
	if id == "" {
		render.SetError(utils.CodeErrNoData, errors.New("任务不存在"))
		return
	}
	step, err := strconv.ParseInt(c.Param("step"), 10, 64)
	if err != nil {
		logrus.Warnln(err)
		render.SetError(utils.CodeErrNoData, errors.New("步骤不存在"))
		return
	}
	taskStepState, found := cache.GetTaskStep(id, step)
	if !found {
		render.SetError(utils.CodeErrNoData, errors.New("步骤不存在"))
		return
	}
	if taskStepState.State == cache.Pending {
		render.SetJson([]string{taskStepState.Message})
		return
	}

	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(c.Request) {
		ws, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
	}
	var latest int64 = -1
	fn := func(latest *int64) []string {
		outputs := cache.GetTaskStepAllOutput(id, step)
		var res []string
		for k, v := range outputs {
			if int64(k) <= *latest {
				continue
			}
			res = append(res, v.Content)
		}
		*latest = int64(len(outputs)) - 1
		return res
	}
	if ws == nil {
		c.Writer.Header().Set("Content-Type", "application/json")
		render.SetJson(fn(&latest))
		return
	}
	defer func() {
		if ws == nil {
			return
		}
		_ = ws.WriteMessage(websocket.CloseMessage, nil)
		_ = ws.Close()
	}()
	for {
		if ws == nil {
			return
		}
		res := fn(&latest)
		if res != nil {
			err = ws.WriteMessage(websocket.TextMessage, []byte(strings.Join(res, "\r\n")))
			if err != nil {
				logrus.Errorln(err)
			}
		}
		if latest <= 0 || cache.GetTaskStepOutputDone(id, step)-2 == latest {
			return
		}
		time.Sleep(time.Millisecond * 30)
	}
}
