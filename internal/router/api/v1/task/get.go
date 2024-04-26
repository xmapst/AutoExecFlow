package task

import (
	"errors"
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
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

// Get
// @Summary task detail
// @description detail task
// @Tags Task
// @Accept application/json
// @Accept application/toml
// @Accept application/x-yaml
// @Accept multipart/form-data
// @Produce application/json
// @Produce application/x-yaml
// @Produce application/toml
// @Param task path string true "task name"
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/task/{task} [get]
func Get(c *gin.Context) {
	render := base.Gin{Context: c}
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(c.Request) {
		var err error
		ws, err = base.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logx.Errorln(err)
			render.SetError(base.CodeNoData, err)
			return
		}
	}

	if ws == nil {
		render.SetNegotiate(get(c))
		return
	}
	// websocket 方式
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		res, err, code := get(c)
		if code == base.CodeNoData {
			err = ws.WriteJSON(base.NewRes(res, err, code))
			if err != nil {
				logx.Errorln(err)
			}
			return
		}
		err = ws.WriteJSON(base.NewRes(res, err, code))
		if err != nil {
			logx.Errorln(err)
			return
		}
	}
}

func get(c *gin.Context) ([]types.StepRes, error, int) {
	taskName := c.Param("task")
	if taskName == "" {
		return nil, errors.New("task does not exist"), base.CodeNoData
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		return nil, err, base.CodeNoData
	}
	state := models.StateMap[*task.State]
	c.Request.Header.Set(types.XTaskState, state)
	c.Writer.Header().Set(types.XTaskState, state)
	c.Set(types.XTaskState, state)
	steps := storage.Task(taskName).StepList("")
	if steps == nil {
		return nil, err, base.CodeNoData
	}
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	var res []types.StepRes
	var groups = make(map[int][]string)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"))
	for _, v := range steps {
		groups[*v.State] = append(groups[*v.State], v.Name)
		res = append(res, procStep(uriPrefix, v))
	}

	var keys []int
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		task.Message = fmt.Sprintf("%s; %s: [%s]", task.Message, models.StateMap[key], strings.Join(groups[key], ","))
	}
	var code int
	switch *task.State {
	case models.Stop:
		code = base.CodeSuccess
	case models.Running:
		code = base.CodeRunning
	case models.Pending:
		code = base.CodePending
	case models.Paused:
		code = base.CodePaused
	case models.Failed:
		code = base.CodeFailed
	default:
		code = base.CodeNoData
	}
	return res, fmt.Errorf(task.Message), code
}

func procStep(uriPrefix string, step *models.Step) types.StepRes {
	res := types.StepRes{
		Name:      step.Name,
		State:     models.StateMap[*step.State],
		Code:      *step.Code,
		Manager:   fmt.Sprintf("%s/step/%s", uriPrefix, step.Name),
		Workspace: fmt.Sprintf("%s/workspace", uriPrefix),
		Timeout:   step.Timeout.String(),
		Env:       make(map[string]string),
		Type:      step.Type,
		Content:   step.Content,
		Time: &types.Time{
			ST: step.STime.Format(time.RFC3339),
			ET: step.ETime.Format(time.RFC3339),
		},
	}

	res.Depends = storage.Task(step.TaskName).Step(step.Name).Depend().List()
	envs := storage.Task(step.TaskName).Step(step.Name).Env().List()
	for _, env := range envs {
		res.Env[env.Name] = env.Value
	}

	var output []string
	if *step.State == models.Stop || *step.State == models.Failed {
		logs := storage.Task(step.TaskName).Step(step.Name).Log().List()
		for _, o := range logs {
			if o.Content == worker.ConsoleStart || o.Content == worker.ConsoleDone {
				continue
			}
			output = append(output, o.Content)
		}
	}

	if output == nil {
		output = []string{step.Message}
	}
	res.Message = strings.Join(output, "\n")
	return res
}
