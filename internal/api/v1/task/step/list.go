package step

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
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

	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now().Add(3*time.Second))
		_ = ws.Close()
	}()

	var ctx, cancel = context.WithCancel(c)
	defer cancel()
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok {
					cancel()
				}
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		code, list, err := service.Task(taskName).Steps()
		if err != nil && code == types.CodeNoData {
			_ = ws.WriteJSON(base.WithCode[any](code).WithError(err))
			return
		}
		err = ws.WriteJSON(base.WithData(list).WithError(err).WithCode(code))
		if err != nil {
			return
		}
		if code == types.CodeSuccess || code == types.CodeFailed {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}
