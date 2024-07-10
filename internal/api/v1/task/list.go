package task

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/api/types"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
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
	page, err := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	if err != nil {
		base.Send(c, types.WithError[any](err))
		return
	}
	size, err := strconv.ParseInt(c.DefaultQuery("size", "100"), 10, 64)
	if err != nil {
		base.Send(c, types.WithError[any](err))
		return
	}
	prefix := c.Query("prefix")
	tasks, total := storage.TaskList(page, size, prefix)
	if tasks == nil {
		base.Send(c, types.WithCode[any](types.CodeSuccess))
		return
	}
	pageTotal := total / size
	if total%size != 0 {
		pageTotal += 1
	}
	var res = &types.TaskListRes{
		Page: types.Page{
			Current: page,
			Size:    size,
			Total:   pageTotal,
		},
	}
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if c.GetHeader("X-Forwarded-Proto") != "" {
		scheme = c.GetHeader("X-Forwarded-Proto")
	}

	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"))
	for _, task := range tasks {
		res.Tasks = append(res.Tasks, procTask(uriPrefix, task))
	}
	base.Send(c, types.WithData(res))
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
		Disable:   *task.Disable,
		Count:     *task.Count,
		Time: &types.Time{
			Start: task.STimeStr(),
			End:   task.ETimeStr(),
		},
	}
	st := storage.Task(task.Name)
	// 获取任务级所有环境变量
	envs := st.Env().List()
	for _, env := range envs {
		res.Env[env.Name] = env.Value
	}

	// 获取当前进行到那些步骤
	steps := st.StepNameList("")
	if steps == nil {
		return res
	}

	var groups = make(map[int][]string)
	for _, name := range steps {
		state, err := st.Step(name).State()
		if err != nil {
			continue
		}
		groups[state] = append(groups[state], name)
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
