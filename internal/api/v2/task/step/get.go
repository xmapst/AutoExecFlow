package step

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

// Detail
// @Summary		Detail
// @description	Get step detail
// @Tags		Step
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Param		step path string true "step name"
// @Success		200 {object} types.Base[types.TaskStepRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v2/task/{task}/step/{step} [get]
func Detail(c *gin.Context) {
	var taskName = c.Param("task")
	if taskName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	var stepName = c.Param("step")
	if stepName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("step does not exist")))
		return
	}
	step, err := storage.Task(taskName).Step(stepName).Get()
	if err != nil {
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
	path := strings.Replace(strings.TrimSuffix(c.Request.URL.Path, "/"), "v2", "v1", 1)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, path)
	data := types.TaskStepRes{
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
	data.Depends = storage.Task(taskName).Step(step.Name).Depend().List()
	envs := storage.Task(taskName).Step(step.Name).Env().List()
	for _, env := range envs {
		data.Env[env.Name] = env.Value
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
	if output != nil {
		data.Message = strings.Join(output, "\n")
	}

	res := types.WithData(data)
	switch *step.State {
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
	base.Send(c, res.WithError(fmt.Errorf(step.Message)))
}
