package step

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Log(w http.ResponseWriter, r *http.Request) {
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(r) {
		var err error
		ws, err = base.Upgrade(w, r)
		if err != nil {
			logx.Errorln(err)
			render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
			return
		}
	}
	var latest int64 = -1
	if ws != nil {
		// 使用websocket方式
		stepDetailWebsocket(r, ws, &latest)
		return
	}
	var taskName = chi.URLParam(r, "task")
	if taskName == "" {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	var stepName = chi.URLParam(r, "step")
	if stepName == "" {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
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
		code = types.CodePending
	case models.Paused:
		res = []*types.LogRes{
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
	render.JSON(w, r, types.New().WithCode(code).WithData(res).WithError(errors.New(step.Message)))
}

func stepDetailWebsocket(r *http.Request, ws *websocket.Conn, latest *int64) {
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	var taskName = chi.URLParam(r, "task")
	if taskName == "" {
		err := ws.WriteJSON(types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		if err != nil {
			logx.Errorln(err)
		}
		return
	}
	var stepName = chi.URLParam(r, "step")
	if stepName == "" {
		err := ws.WriteJSON(types.New().WithCode(types.CodeNoData).WithError(errors.New("step does not exist")))
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
			err = ws.WriteJSON(types.New().WithCode(types.CodeNoData).WithError(err))
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
				err = ws.WriteJSON(types.New().WithCode(types.CodePending).WithData([]types.LogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      1,
						Content:   step.Message,
					},
				}))
				if err != nil {
					logx.Errorln(err)
					return
				}
			})
			continue
		case models.Paused:
			// 只发送一次
			pausedOnce.Do(func() {
				err = ws.WriteJSON(types.New().WithCode(types.CodePaused).WithData([]types.LogRes{
					{
						Timestamp: time.Now().UnixNano(),
						Line:      1,
						Content:   "step is paused",
					},
				}))
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
				err = ws.WriteJSON(types.New().WithCode(types.CodeFailed).WithData(res).WithError(errMsg))
				if err != nil {
					logx.Errorln(err)
				}
				return
			}
			err = ws.WriteJSON(types.New().WithData(res))
			if err != nil {
				logx.Errorln(err)
			}
			return
		}
	}

	for range ticker.C {
		res, done := getTaskStepLog(taskName, stepName, latest)
		for _, v := range res {
			if err := ws.WriteJSON(types.New().WithCode(types.CodeRunning).WithData(v).WithError(errors.New("in progress"))); err != nil {
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
