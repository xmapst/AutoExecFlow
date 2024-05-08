package task

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/logx"
)

func List(c *gin.Context) {
	render := base.Gin{Context: c}
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(c.Request) {
		var err error
		ws, err = base.Upgrade(c.Writer, c.Request)
		if err != nil {
			logx.Errorln(err)
			render.SetNegotiate(&types.TaskListRes{}, err, base.CodeSuccess)
			return
		}
	}
	if ws == nil {
		render.SetRes(list(c))
		return
	}
	// websocket 方式
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	// 使用websocket方式
	var ticker = time.NewTicker(1 * time.Second)
	for range ticker.C {
		err := ws.WriteJSON(base.NewRes(list(c), nil, base.CodeSuccess))
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
}

// 每次获取全量数据
func list(c *gin.Context) *types.TaskListRes {
	tasks := storage.TaskList("")
	if tasks == nil {
		return nil
	}
	var res = &types.TaskListRes{
		Total: len(tasks),
	}
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"))
	for _, task := range tasks {
		res.Tasks = append(res.Tasks, procTask(uriPrefix, task))
	}
	return res
}

func procTask(uriPrefix string, task *models.Task) *types.TaskRes {
	res := &types.TaskRes{
		Name:      task.Name,
		State:     models.StateMap[*task.State],
		Manager:   fmt.Sprintf("%s/%s", uriPrefix, task.Name),
		Workspace: fmt.Sprintf("%s/%s/workspace", uriPrefix, task.Name),
		Message:   task.Message,
		Env:       make(map[string]string),
		Timeout:   task.Timeout.String(),
		Count:     *task.Count,
		Time: &types.Time{
			Start: task.STimeStr(),
			End:   task.ETimeStr(),
		},
	}

	// 获取任务级所有环境变量
	envs := storage.Task(task.Name).Env().List()
	for _, env := range envs {
		res.Env[env.Name] = env.Value
	}

	// 获取当前进行到那些步骤
	steps := storage.Task(task.Name).StepList("")
	if steps == nil {
		return res
	}

	var groups = make(map[int][]string)
	for _, vv := range steps {
		groups[*vv.State] = append(groups[*vv.State], vv.Name)
	}
	var keys []int
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		res.Message = fmt.Sprintf("%s; %s: [%s]", res.Message, models.StateMap[key], strings.Join(groups[key], ","))
	}
	return res
}
