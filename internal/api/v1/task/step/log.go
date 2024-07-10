package step

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

// Log
// @Summary		Log
// @description	Step Execution Output
// @Tags		Step
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Param		step path string true "step name"
// @Success		200 {object} types.Base[[]types.TaskStepLogRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task/{task}/step/{step} [get]
func Log(c *gin.Context) {
	var ws *websocket.Conn
	if c.IsWebsocket() {
		var err error
		ws, err = base.Upgrade(c.Writer, c.Request)
		if err != nil {
			logx.Errorln(err)
			base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
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
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	var code int
	var res []*types.TaskStepLogRes
	switch *step.State {
	case models.Pending:
		res = []*types.TaskStepLogRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   step.Message,
			},
		}
		code = types.CodePending
	case models.Paused:
		res = []*types.TaskStepLogRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   "step is paused",
			},
		}
		code = types.CodePaused
	default:
		res, _ = getTaskStepLog(taskName, stepName, &latest)
		switch *step.State {
		case models.Stop:
			code = types.CodeSuccess
		case models.Running:
			code = types.CodeRunning
		case models.Failed:
			code = types.CodeFailed
		default:
			code = types.CodeNoData
		}
	}
	base.Send(c, types.WithData(res).WithCode(code).WithError(errors.New(step.Message)))
}

func stepDetailWebsocket(c *gin.Context, ws *websocket.Conn, latest *int64) {
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	var taskName = c.Param("task")
	if taskName == "" {
		err := ws.WriteJSON(types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		if err != nil {
			logx.Errorln(err)
		}
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		err := ws.WriteJSON(types.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
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
			err = ws.WriteJSON(types.WithCode[any](types.CodeNoData).WithError(err))
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
				err = ws.WriteJSON(types.WithData([]types.TaskStepLogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      1,
						Content:   step.Message,
					},
				}).WithCode(types.CodePending))
				if err != nil {
					logx.Errorln(err)
					return
				}
			})
			continue
		case models.Paused:
			// 只发送一次
			pausedOnce.Do(func() {
				err = ws.WriteJSON(types.WithData([]types.TaskStepLogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      1,
						Content:   "step is paused",
					},
				}).WithCode(types.CodePaused))
				if err != nil {
					logx.Errorln(err)
					return
				}
			})
			continue
		default:
			res, _ := getTaskStepLog(taskName, stepName, latest)
			if res == nil {
				res = []*types.TaskStepLogRes{
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
				err = ws.WriteJSON(types.WithData(res).WithCode(types.CodeFailed).WithError(errMsg))
				if err != nil {
					logx.Errorln(err)
				}
				return
			}
			err = ws.WriteJSON(types.WithData(res).WithCode(types.CodeSuccess))
			if err != nil {
				logx.Errorln(err)
			}
			return
		}
	}

	for range ticker.C {
		res, done := getTaskStepLog(taskName, stepName, latest)
		for _, v := range res {
			if err := ws.WriteJSON(types.WithData(v).WithCode(types.CodeRunning).WithError(errors.New("in progress"))); err != nil {
				logx.Errorln(err)
				return
			}
		}

		if *latest <= 0 || done {
			return
		}
	}
}

func getTaskStepLog(task, step string, latest *int64) (res []*types.TaskStepLogRes, done bool) {
	logs := storage.Task(task).Step(step).Log().List()
	for k, v := range logs {
		if v.Content == common.ConsoleStart {
			continue
		}
		if v.Content == common.ConsoleDone {
			done = true
			continue
		}
		if int64(k) <= *latest {
			continue
		}
		res = append(res, &types.TaskStepLogRes{
			Timestamp: v.Timestamp,
			Line:      *v.Line,
			Content:   v.Content,
		})
	}
	*latest = int64(len(logs)) - 1
	return
}
