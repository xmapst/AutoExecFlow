package step

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

// Log
// @Summary task step log
// @description detail task step log
// @Tags Task
// @Accept application/json
// @Accept application/toml
// @Accept application/x-yaml
// @Accept multipart/form-data
// @Produce application/json
// @Produce application/x-yaml
// @Produce application/toml
// @Param task path string true "task name"
// @Param step path string true "step name"
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task/{task}/step/{step} [get]
func Log(c *gin.Context) {
	var render = base.Gin{Context: c}
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(c.Request) {
		var err error
		ws, err = base.Upgrade(c.Writer, c.Request)
		if err != nil {
			logx.Errorln(err)
			render.SetError(base.CodeNoData, err)
			return
		}
	}
	var latest int64 = -1
	if ws != nil {
		// 使用websocket方式
		stepDetailWebsocket(c, ws, &latest)
		return
	}
	var taskName = c.Param("task")
	if taskName == "" {
		render.SetError(base.CodeNoData, errors.New("task does not exist"))
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		render.SetError(base.CodeNoData, errors.New("step does not exist"))
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		render.SetError(base.CodeNoData, err)
		return
	}
	var code int
	var res []*types.LogRes
	switch *step.State {
	case models.Pending:
		res = []*types.LogRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   step.Message,
			},
		}
		code = base.CodePending
	case models.Paused:
		res = []*types.LogRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   "step is paused",
			},
		}
		code = base.CodePaused
	default:
		res, _ = getTaskStepLog(taskName, stepName, &latest)
		switch *step.State {
		case models.Stop:
			code = base.CodeSuccess
		case models.Running:
			code = base.CodeRunning
		case models.Failed:
			code = base.CodeFailed
		default:
			code = base.CodeNoData
		}
	}
	render.SetNegotiate(res, errors.New(step.Message), code)
}

func stepDetailWebsocket(c *gin.Context, ws *websocket.Conn, latest *int64) {
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	var taskName = c.Param("task")
	if taskName == "" {
		err := ws.WriteJSON(base.NewRes(nil, errors.New("task does not exist"), base.CodeNoData))
		if err != nil {
			logx.Errorln(err)
		}
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		err := ws.WriteJSON(base.NewRes(nil, errors.New("step does not exist"), base.CodeNoData))
		if err != nil {
			logx.Errorln(err)
		}
		return
	}

	var ticker = time.NewTicker(1 * time.Second)
	var pendingOnce sync.Once
	var pausedOnce sync.Once
	for range ticker.C {
		step, err := storage.Task(taskName).Step(stepName).Get()
		if err != nil {
			err = ws.WriteJSON(base.NewRes(nil, err, base.CodeNoData))
			if err != nil {
				logx.Errorln(err)
			}
			return
		}
		if *step.State == models.Running {
			break
		}
		switch *step.State {
		case models.Pending:
			// 只发送一次
			pendingOnce.Do(func() {
				err = ws.WriteJSON(base.NewRes([]types.LogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      1,
						Content:   step.Message,
					},
				}, nil, base.CodePending))
				if err != nil {
					logx.Errorln(err)
					return
				}
			})
			continue
		case models.Paused:
			// 只发送一次
			pausedOnce.Do(func() {
				err = ws.WriteJSON(base.NewRes([]types.LogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      1,
						Content:   "step is paused",
					},
				}, nil, base.CodePaused))
				if err != nil {
					logx.Errorln(err)
					return
				}
			})
			continue
		default:
			res, _ := getTaskStepLog(taskName, stepName, latest)
			if res == nil {
				res = []*types.LogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      0,
						Content:   step.Message,
					},
				}
			}
			if *step.Code != 0 {
				errMsg := fmt.Errorf("exit code: %d", step.Code)
				if step.Message != "" {
					errMsg = fmt.Errorf(step.Message)
				}
				err = ws.WriteJSON(base.NewRes(res, errMsg, base.CodeFailed))
				if err != nil {
					logx.Errorln(err)
				}
				return
			}
			err = ws.WriteJSON(base.NewRes(res, nil, base.CodeSuccess))
			if err != nil {
				logx.Errorln(err)
			}
			return
		}
	}

	for range ticker.C {
		res, done := getTaskStepLog(taskName, stepName, latest)
		for _, v := range res {
			if err := ws.WriteJSON(base.NewRes(v, errors.New("in progress"), base.CodeRunning)); err != nil {
				logx.Errorln(err)
				return
			}
		}

		if *latest <= 0 || done {
			return
		}
	}
}

func getTaskStepLog(task, step string, latest *int64) (res []*types.LogRes, done bool) {
	logs := storage.Task(task).Step(step).Log().List()
	for k, v := range logs {
		if v.Content == worker.ConsoleStart {
			continue
		}
		if v.Content == worker.ConsoleDone {
			done = true
			continue
		}
		if int64(k) <= *latest {
			continue
		}
		res = append(res, &types.LogRes{
			Timestamp: v.Timestamp,
			Line:      *v.Line,
			Content:   v.Content,
		})
	}
	*latest = int64(len(logs)) - 1
	return
}
