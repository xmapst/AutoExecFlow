package step

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
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
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	task, err := storage.Task(taskName).Get()
	if err != nil {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	state := models.StateMap[*task.State]
	c.Request.Header.Set(types.XTaskState, state)
	c.Header(types.XTaskState, state)
	steps := storage.Task(taskName).StepList(backend.All)
	if steps == nil {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if c.GetHeader("X-Forwarded-Proto") != "" {
		scheme = c.GetHeader("X-Forwarded-Proto")
	}

	var data []types.TaskStepRes
	var groups = make(map[int][]string)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"))
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

	res := types.WithData(data)
	switch *task.State {
	case models.Stop:
		res.WithCode(types.CodeSuccess)
	case models.Running:
		res.WithCode(types.CodeRunning)
	case models.Pending:
		res.WithCode(types.CodePending)
	case models.Paused:
		res.WithCode(types.CodePaused)
	case models.Failed:
		res.WithCode(types.CodeFailed)
	default:
		res.WithCode(types.CodeNoData)
	}
	base.Send(c, res.WithError(fmt.Errorf(task.Message)))
}

func procStep(uriPrefix string, taskName string, step *models.Step) types.TaskStepRes {
	res := types.TaskStepRes{
		Name:      step.Name,
		State:     models.StateMap[*step.State],
		Code:      *step.Code,
		Manager:   fmt.Sprintf("%s/step/%s", uriPrefix, step.Name),
		Workspace: fmt.Sprintf("%s/workspace", uriPrefix),
		Message:   step.Message,
		Timeout:   step.Timeout.String(),
		Disable:   *step.Disable,
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
			if o.Content == common.ConsoleStart || o.Content == common.ConsoleDone {
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
