package task

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
)

func List(w http.ResponseWriter, r *http.Request) {
	tasks := storage.TaskList("")
	if tasks == nil {
		base.SendJson(w, base.New())
		return
	}
	var res = &types.TaskListRes{
		Total: len(tasks),
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
