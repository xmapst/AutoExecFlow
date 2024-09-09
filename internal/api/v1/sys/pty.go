package sys

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
	_ "github.com/xmapst/AutoExecFlow/types"
)

// PtyWs
// @Summary		Terminal
// @description	websocket terminal
// @Tags		System
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Success		200 {object} types.Base[any]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/pty [get]
func PtyWs(c *gin.Context) {
	ws, err := base.Upgrade(c.Writer, c.Request)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	defer ws.Close()

	pty, err := service.Pty(ws)
	if err != nil {
		logx.Errorln(err)
		_ = ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), time.Now().Add(time.Second))
		return
	}
	pty.Run()
}
