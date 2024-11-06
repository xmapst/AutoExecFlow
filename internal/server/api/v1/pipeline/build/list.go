package build

import (
	"context"
	"encoding/json"
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
// @Summary		列表
// @Description	获取指定流水线构建任务列表, 支持WS长连接
// @Tags		构建
// @Accept		application/json
// @Produce		application/json
// @Param		pipeline path string true "流水线名称"
// @Param		page query int false "页码" default(1)
// @Param		size query int false "分页大小" default(100)
// @Success		200 {object} types.SBase[types.SPipelineBuildListRes]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/pipeline/{pipeline}/build [get]
func List(c *gin.Context) {
	pipelineName := c.Param("pipeline")
	if pipelineName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("pipeline does not exist")))
		return
	}

	var req = &types.SPageReq{
		Page: 1,
		Size: 15,
	}
	if err := c.ShouldBindQuery(req); err != nil {
		base.Send(c, base.WithError[any](err))
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
		list := service.Pipeline(pipelineName).BuildList(req)
		base.Send(c, base.WithData(list))
		return
	}

	defer base.CloseWs(ws, "Server is shutting down")

	var ctx, cancel = context.WithCancel(c)
	defer cancel()
	go func() {
		for {
			t, p, err := ws.ReadMessage()
			if err != nil {
				cancel()
				return
			}
			switch t {
			case websocket.TextMessage:
				err = json.Unmarshal(p, &req)
				if err != nil {
					continue
				}
			}
		}
	}()

	var lastPipelineList *types.SPipelineBuildListRes // 缓存上一次的推送数据
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		currentPipelineList := service.Pipeline(pipelineName).BuildList(req)
		// 如果数据没有变化，只发送心跳
		if reflect.DeepEqual(lastPipelineList, currentPipelineList) {
			err := ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		err := ws.WriteJSON(base.WithData(currentPipelineList))
		if err != nil {
			return
		}
		// 保存当前数据作为上一次的数据
		lastPipelineList = currentPipelineList
		time.Sleep(300 * time.Millisecond)
	}
}
