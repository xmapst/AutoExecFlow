package task

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/internal/worker"
)

func Get(w http.ResponseWriter, r *http.Request) {
	taskName := chi.URLParam(r, "task")
	if taskName == "" {
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
		return
	}
	state := models.StateMap[*task.State]
	r.Header.Set(types.XTaskState, state)
	w.Header().Set(types.XTaskState, state)
	steps := storage.Task(taskName).StepList(backend.All)
	if steps == nil {
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(err))
		return
	}
	var scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	var data []types.StepRes
	var groups = make(map[int][]string)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, r.Host, strings.TrimSuffix(r.URL.Path, "/"))
	for _, v := range steps {
		groups[*v.State] = append(groups[*v.State], v.Name)
		data = append(data, procStep(uriPrefix, taskName, v))
	}

	var keys []int
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		task.Message = fmt.Sprintf("%s; %s: [%s]", task.Message, models.StateMap[key], strings.Join(groups[key], ","))
	}

	res := base.New().WithData(data)
	switch *task.State {
	case models.Stop:
		res.WithCode(base.CodeSuccess)
	case models.Running:
		res.WithCode(base.CodeRunning)
	case models.Pending:
		res.WithCode(base.CodePending)
	case models.Paused:
		res.WithCode(base.CodePaused)
	case models.Failed:
		res.WithCode(base.CodeFailed)
	default:
		res.WithCode(base.CodeNoData)
	}
	base.SendJson(w, res.WithError(fmt.Errorf(task.Message)))
}

func procStep(uriPrefix string, taskName string, step *models.Step) types.StepRes {
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

	res.Depends = storage.Task(taskName).Step(step.Name).Depend().List()
	envs := storage.Task(taskName).Step(step.Name).Env().List()
	for _, env := range envs {
		res.Env[env.Name] = env.Value
	}

	var output []string
	if *step.State == models.Stop || *step.State == models.Failed {
		logs := storage.Task(taskName).Step(step.Name).Log().List()
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
