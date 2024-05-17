package task

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
)

func List(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.ParseInt(base.DefaultQuery(r, "page", "1"), 10, 64)
	if err != nil {
		base.SendJson(w, base.New().WithError(err))
		return
	}
	size, err := strconv.ParseInt(base.DefaultQuery(r, "size", "100"), 10, 64)
	if err != nil {
		base.SendJson(w, base.New().WithError(err))
		return
	}
	prefix := base.Query(r, "prefix")
	tasks, total := storage.TaskList(page, size, prefix)
	if tasks == nil {
		base.SendJson(w, base.New())
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
	if r.TLS != nil {
		scheme = "https"
	}
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, r.Host, strings.TrimSuffix(r.URL.Path, "/"))
	for _, task := range tasks {
		res.Tasks = append(res.Tasks, procTask(uriPrefix, task))
	}
	base.SendJson(w, base.New().WithData(res))
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
