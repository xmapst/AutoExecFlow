package pty

import (
	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// PtyWs
// @Summary		PTY
// @Description	Websocket的pty终端
// @Tags		终端
// @Accept		application/json
// @Produce		application/json
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pty [get]
func Websocket(c *gin.Context) {
	ws, err := base.Upgrade(c.Writer, c.Request)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}

	pty, err := service.Pty(ws)
	if err != nil {
		base.CloseWs(ws, err.Error())
		return
	}
	pty.Run()
	base.CloseWs(ws, "Server is shutting down")
}
