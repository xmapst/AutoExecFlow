package step

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/types"
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
	taskName := c.Param("task")
	stepName := c.Param("step")
	if taskName == "" || stepName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task or step does not exist")))
		return
	}

	var ws *websocket.Conn
	if c.IsWebsocket() {
		var err error
		ws, err = base.Upgrade(c.Writer, c.Request)
		if err != nil {
			base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
			return
		}
	}
	if ws != nil {
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
		defer base.CloseWs(ws, "Server is shutting down")
		err := service.Step(taskName, stepName).LogStream(ctx, ws)
		if err != nil {
			_ = ws.WriteJSON(base.WithCode[any](types.CodeFailed).WithError(err))
		}
		return
	}

	code, res, err := service.Step(taskName, stepName).Log()
	base.Send(c, base.WithData(res).WithCode(code).WithError(err))
}
