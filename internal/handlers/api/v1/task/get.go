package task

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/handlers/types"
	"github.com/xmapst/osreapi/internal/logx"
	"github.com/xmapst/osreapi/internal/utils"
)

// Detail
// @Summary task detail
// @description detail task
// @Tags Task
// @Accept json
// @Produce json
// @Param task path string true "task id"
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task/{task} [get]
func Detail(c *gin.Context) {
	render := base.Gin{Context: c}
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeErrNoData, errors.New("task does not exist"))
		return
	}
	taskState, found := cache.GetTask(task)
	if !found {
		render.SetError(base.CodeErrNoData, errors.New("task does not exist"))
		return
	}
	state := cache.StateMap[taskState.State]
	c.Request.Header.Set(types.XTaskState, state)
	c.Writer.Header().Set(types.XTaskState, state)
	c.Set(types.XTaskState, state)
	tasksStepStates := cache.GetTaskStepStates(task)
	var res []types.TaskDetailRes
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	var code int64
	var stopMsg, runMsg []string
	for _, v := range tasksStepStates {
		code += v.Code
		msg := v.Message
		if v.Code != 0 {
			msg = fmt.Sprintf("Step: %d, Exit Code: %d", v.ID, v.Code)
			if v.Name != "" {
				msg = fmt.Sprintf("Step: %d, Name: %s, Exit Code: %d", v.ID, v.Name, v.Code)
			}
			if taskState.MetaData.VMInstanceID != "" {
				msg += fmt.Sprintf(", Instance ID: %s", taskState.MetaData.VMInstanceID)
			}
			if taskState.MetaData.HardWareID != "" {
				msg += fmt.Sprintf(", Hardware ID: %s", taskState.MetaData.HardWareID)
			}
			stopMsg = append(stopMsg, msg)
		}
		var output []string
		if v.State == cache.Running {
			msg = fmt.Sprintf("Step: %d, Name: %s", v.ID, v.Name)
			if v.Name == "" {
				msg = fmt.Sprintf("Step: %d", v.ID)
			}
			runMsg = append(runMsg, msg)
			output = []string{"The step is running"}
		}
		if v.State == cache.Stop {
			outputs := cache.GetTaskStepAllOutput(task, v.ID)
			for _, o := range outputs {
				output = append(output, o.Content)
			}
		}
		if output == nil {
			output = []string{v.Message}
		}
		_res := types.TaskDetailRes{
			ID:        v.ID,
			URL:       fmt.Sprintf("%s://%s%s/%d/console", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), v.ID),
			Name:      v.Name,
			State:     v.State,
			Code:      v.Code,
			DependsOn: v.DependsOn,
			Message:   strings.Join(output, "\n"),
			Times: &types.Times{
				ST: utils.TimeStr(v.Times.ST),
				ET: utils.TimeStr(v.Times.ET),
				RT: v.Times.RT.String(),
			},
		}
		res = append(res, _res)
	}

	switch taskState.State {
	// 运行结束
	case cache.Stop:
		if code != 0 {
			render.SetRes(res, fmt.Errorf("[%s]", strings.Join(stopMsg, "; ")), base.CodeExecErr)
			return
		}
		render.SetJson(res)
	// 运行中, 排队中
	case cache.Running:
		render.SetRes(res, fmt.Errorf("[%s]", strings.Join(runMsg, "; ")), base.CodeRunning)
	case cache.Pending:
		render.SetError(base.CodeRunning, nil)
	case cache.SystemError:
		render.SetRes(res, fmt.Errorf(taskState.Message), base.CodeExecErr)
	default:
		render.SetError(base.CodeErrNoData, errors.New("task does not exist"))
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// StepDetail
// @Summary task step detail
// @description detail task step
// @Tags Task
// @Accept json
// @Produce json
// @Param task path string true "task id"
// @Param step path string true "step id" default(0)
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task/{task}/{step}/console [get]
func StepDetail(c *gin.Context) {
	render := base.Gin{Context: c}
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeErrNoData, errors.New("task does not exist"))
		return
	}
	step, err := strconv.ParseInt(c.Param("step"), 10, 64)
	if err != nil {
		logx.Warningln(err)
		render.SetError(base.CodeErrNoData, errors.New("step does not exist"))
		return
	}
	taskStepState, found := cache.GetTaskStepState(task, step)
	if !found {
		render.SetError(base.CodeErrNoData, errors.New("step does not exist"))
		return
	}
	if taskStepState.State == cache.Pending {
		render.SetRes([]string{taskStepState.Message}, nil, base.CodePending)
		return
	}

	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(c.Request) {
		ws, err = upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
	var latest int64 = -1
	fn := func(latest *int64) []string {
		outputs := cache.GetTaskStepAllOutput(task, step)
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
		if taskStepState.Code != 0 {
			res := fn(&latest)
			if res == nil {
				res = []string{taskStepState.Message}
			}
			render.SetRes(res,
				fmt.Errorf("exit code: %d; %s", taskStepState.Code, taskStepState.Message),
				base.CodeExecErr,
			)
			return
		}
		if taskStepState.State == cache.Running {
			render.SetRes(fn(&latest), errors.New("in progress"), base.CodeRunning)
			return
		}
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
		res := fn(&latest)
		if res != nil {
			bytes := []byte(strings.Join(res, "\r\n"))
			err = ws.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				logx.Errorln(err)
			}
		}
		if latest <= 0 || cache.GetTaskStepOutputDone(task, step)-2 == latest {
			return
		}
		time.Sleep(time.Millisecond * 30)
	}
}
