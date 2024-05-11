package task

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/worker"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Get(w http.ResponseWriter, r *http.Request) {
	var ws *websocket.Conn
	if websocket.IsWebSocketUpgrade(r) {
		var err error
		ws, err = base.Upgrade(w, r)
		if err != nil {
			logx.Errorln(err)
			base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
			return
		}
	}

	if ws == nil {
		res, err, code := get(w, r)
		base.SendJson(w, base.New().WithCode(code).WithData(res).WithError(err))
		return
	}
	// websocket 方式
	defer func() {
		_ = ws.WriteControl(websocket.CloseMessage, nil, time.Now())
		_ = ws.Close()
	}()
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		res, err, code := get(w, r)
		err = ws.WriteJSON(base.New().WithCode(code).WithData(res).WithError(err))
		if err != nil {
			logx.Errorln(err)
		}
		if code == base.CodeNoData {
			return
		}
	}
}

func get(w http.ResponseWriter, r *http.Request) ([]types.StepRes, error, int) {
	taskName := chi.URLParam(r, "task")
	if taskName == "" {
		return nil, errors.New("task does not exist"), base.CodeNoData
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		return nil, err, base.CodeNoData
	}
	state := models.StateMap[*task.State]
	r.Header.Set(types.XTaskState, state)
	w.Header().Set(types.XTaskState, state)
	steps := storage.Task(taskName).StepList("")
	if steps == nil {
		return nil, err, base.CodeNoData
	}
	var scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	var res []types.StepRes
	var groups = make(map[int][]string)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, r.Host, strings.TrimSuffix(r.URL.Path, "/"))
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
			Start: step.STimeStr(),
			End:   step.ETimeStr(),
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
