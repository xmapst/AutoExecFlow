package task

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/service"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// List
// @Summary		List
// @description	Get the all task list
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		page query int false "page number" default(1)
// @Param		size query int false "paging Size" default(100)
// @Param		prefix query string false "Keywords"
// @Success		200 {object} types.Base[types.TaskListRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task [get]
func List(c *gin.Context) {
	var req = &types.PageReq{
		Page: 1,
		Size: 10,
	}
	err := c.ShouldBindQuery(req)
	if err != nil {
		base.Send(c, base.WithError[any](err))
		return
	}
	var ws *websocket.Conn
	if c.IsWebsocket() {
		ws, err = base.Upgrade(c.Writer, c.Request)
		if err != nil {
			logx.Errorln(err)
			base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
			return
		}
	}

	if ws == nil {
		list := service.TaskList(req)
		base.Send(c, base.WithData(list))
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
			t, p, err := ws.ReadMessage()
			if err != nil {
				if _, ok := err.(*websocket.CloseError); ok {
					cancel()
				}
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

	var lastTaskList *types.TaskListRes // 缓存上一次的推送数据
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		currentTaskList := service.TaskList(req)
		// 如果数据没有变化，只发送心跳
		if reflect.DeepEqual(lastTaskList, currentTaskList) {
			err = ws.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		err = ws.WriteJSON(base.WithData(currentTaskList))
		if err != nil {
			return
		}
		// 保存当前数据作为上一次的数据
		lastTaskList = currentTaskList
		time.Sleep(300 * time.Millisecond)
	}
}
