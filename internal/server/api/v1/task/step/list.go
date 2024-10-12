package step

import (
	"context"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// List
// @Summary		List
// @description	Get the list of steps for a specified task
// @Tags		Step
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Success		200 {object} types.Base[[]types.TaskStepRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task/{task} [get]
func List(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}

	var ws *websocket.Conn
	if c.IsWebsocket() {
		var err error
		ws, err = base.Upgrade(c.Writer, c.Request)
		if err != nil {
			logx.Errorln(err)
			base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
			return
		}
	}

	if ws == nil {
		code, list, err := service.Task(taskName).Steps()
		base.Send(c, base.WithData(list).WithError(err).WithCode(code))
		return
	}

	defer base.CloseWs(ws, "Server is shutting down")

	var ctx, cancel = context.WithCancel(c)
	defer cancel()
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				cancel()
				return
			}
		}
	}()
	var lastCode types.Code
	var lastError error
	var lastList []*types.TaskStepRes
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		currentCode, currentList, currentErr := service.Task(taskName).Steps()
		if currentErr != nil && currentCode == types.CodeNoData {
			_ = ws.WriteJSON(base.WithCode[any](currentCode).WithError(currentErr))
			return
		}
		// 如果状态, 错误, 列表均没有变化, 则只发送心跳
		if lastCode == currentCode && errors.Is(currentErr, lastError) && reflect.DeepEqual(lastList, currentList) {
			err := ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// 保存当前数据作为上一次的数据
		lastCode = currentCode
		lastError = currentErr
		lastList = currentList

		err := ws.WriteJSON(base.WithData(currentList).WithError(currentErr).WithCode(currentCode))
		if err != nil {
			return
		}
		if currentCode == types.CodeSuccess || currentCode == types.CodeFailed {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
