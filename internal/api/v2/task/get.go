package taskv2

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

// Detail
// @Summary		Detail
// @description	Get task detail
// @Tags		Task
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Success		200 {object} types.Base[types.TaskRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v2/task/{task} [get]
func Detail(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	db := storage.Task(taskName)
	task, err := db.Get()
	if err != nil {
		base.Send(c, types.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	state := models.StateMap[*task.State]
	c.Request.Header.Set(types.XTaskState, state)
	c.Header(types.XTaskState, state)

	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if c.GetHeader("X-Forwarded-Proto") != "" {
		scheme = c.GetHeader("X-Forwarded-Proto")
	}
	path := strings.Replace(strings.TrimSuffix(c.Request.URL.Path, "/"), "v2", "v1", 1)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, path)
	data := &types.TaskRes{
		Name:      task.Name,
		State:     models.StateMap[*task.State],
		Manager:   fmt.Sprintf("%s", uriPrefix),
		Workspace: fmt.Sprintf("%s/workspace", uriPrefix),
		Message:   task.Message,
		Env:       make(map[string]string),
		Timeout:   task.Timeout.String(),
		Disable:   *task.Disable,
		Count:     *task.Count,
		Time: &types.Time{
			Start: task.STimeStr(),
			End:   task.ETimeStr(),
		},
	}
	for _, env := range db.Env().List() {
		data.Env[env.Name] = env.Value
	}

	// 获取当前进行到那些步骤
	steps := db.StepNameList("")
	if steps != nil {
		var groups = make(map[int][]string)
		for _, name := range steps {
			_state, err := db.Step(name).State()
			if err != nil {
				continue
			}
			groups[_state] = append(groups[_state], name)
		}
		var keys []int
		for k := range groups {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, key := range keys {
			data.Message = fmt.Sprintf("%s; %s: [%s]", data.Message, models.StateMap[key], strings.Join(groups[key], ","))
		}
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
