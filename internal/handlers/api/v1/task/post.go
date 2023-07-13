package task

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"github.com/segmentio/ksuid"
	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/dag"
	"github.com/xmapst/osreapi/internal/engine/manager"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/logx"
	"github.com/xmapst/osreapi/internal/utils"
)

type Request struct {
	CommandType    string   `json:"command_type" form:"command_type" binding:"required" example:"powershell"`
	CommandContent string   `json:"command_content" form:"command_content" binding:"required" example:"sleep 10"`
	Name           string   `json:"name,omitempty" form:"name,omitempty" example:"script.ps1"`
	EnvVars        []string `json:"env_vars,omitempty" form:"env_vars,omitempty" example:"env1=value1,env2=value2"`
	DependsOn      []string `json:"depends_on,omitempty" form:"depends_on,omitempty"  example:""`
	TimeOut        string   `json:"time_out,omitempty" form:"time_out,omitempty" example:"3m"`
	timeout        time.Duration
}

type Requests []Request

// Post
// @Summary post task
// @description post task step
// @Tags Task
// @Accept json
// @Produce json
// @param name query string false "task name"
// @Param ansync query bool false "task asynchronously" default(false)
// @Param scripts body Requests true "scripts"
// @Success 200 {object} base.Result
// @Failure 500 {object} base.Result
// @Router /api/v1/task [post]
func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	name := c.Query("name")
	ansync := c.DefaultQuery("ansync", "false")
	var req Requests
	if err := c.ShouldBind(&req); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}
	if req == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   base.GetMsg(base.CodeErrParam),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}

	var task = cache.Task{
		ID: ksuid.New().String(),
	}
	if name != "" {
		if manager.TaskRunning(name) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":      http.StatusBadRequest,
				"message":   "task is running",
				"timestamp": time.Now().UnixNano(),
			})
			return
		}
		task.ID = name
	}
	req.fixName(task.ID)
	if ansync == "false" {
		req.fixSync()
	}

	// Check the uniqueness of the name
	if err := req.uniqNames(); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}
	req.parseDuration()

	for _, v := range req {
		task.Steps = append(task.Steps, &cache.TaskStep{
			Name:           v.Name,
			CommandType:    v.CommandType,
			CommandContent: v.CommandContent,
			EnvVars:        v.EnvVars,
			DependsOn:      v.DependsOn,
			Timeout:        v.timeout,
		})
	}

	// 检查是否回环
	if err := checkFlow(task); err != nil {
		logx.Errorln(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":      http.StatusBadRequest,
			"message":   err.Error(),
			"timestamp": time.Now().UnixNano(),
		})
		return
	}

	task.MetaData = req.getHardWareIDAndVmInstanceID()
	c.Request.Header.Set(xTaskID, task.ID)
	c.Writer.Header().Set(xTaskID, task.ID)
	c.Set(xTaskID, task.ID)

	// 加入池中异步处理
	worker.Submit(task)
	var scheme = "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	render.SetJson(gin.H{
		"count":     len(req),
		"id":        task.ID,
		"url":       fmt.Sprintf("%s://%s%s/%s", scheme, c.Request.Host, strings.TrimSuffix(c.Request.URL.Path, "/"), task.ID),
		"timestamp": time.Now().UnixNano(),
	})
}

func (p Requests) fixName(taskID string) {
	for step, v := range p {
		if v.Name == "" {
			p[step].Name = fmt.Sprintf("%s-%d", taskID, step)
		}
	}
}

func (p Requests) fixSync() {
	for k := range p {
		if k == 0 {
			p[k].DependsOn = nil
			continue
		}
		p[k].DependsOn = []string{p[k-1].Name}
	}
}

func (p Requests) uniqNames() (result error) {
	counts := make(map[string]int)
	for _, v := range p {
		counts[v.Name]++
	}
	for name, count := range counts {
		if count > 1 {
			result = multierror.Append(result, fmt.Errorf("%s repeat count %d", name, count))
		}
	}
	return
}

func (p Requests) parseDuration() {
	for k, v := range p {
		timeout, err := time.ParseDuration(v.TimeOut)
		if v.TimeOut == "" || err != nil {
			timeout = config.App.ExecTimeOut
		}
		p[k].timeout = timeout
	}
}

func (p Requests) getHardWareIDAndVmInstanceID() (res cache.MetaData) {
	// check envs
	for k, v := range p {
		var _env []string
		m := utils.SliceToStrMap(v.EnvVars)
		for k, v := range m {
			_env = append(_env, fmt.Sprintf("%s=%s", k, v))
		}
		str, ok := m["HARDWARE_ID"]
		if ok && str != "" {
			res.HardWareID = str
		}
		str, ok = m["VM_INSTANCE_ID"]
		if ok && str != "" {
			res.VMInstanceID = str
		}
		p[k].EnvVars = _env
	}
	return
}

func checkFlow(task cache.Task) error {
	var stepFnMap = make(map[string]*dag.Step)
	for _, v := range task.Steps {
		step := v
		fn := func(ctx context.Context) error { return nil }
		stepFnMap[step.Name] = dag.NewStep(step.Name, fn)
	}

	// 编排步骤: 创建一个有向无环图，图中的每个顶点都是一个作业
	var flow = dag.NewTask()
	for _, step := range task.Steps {
		stepFn, ok := stepFnMap[step.Name]
		if !ok {
			continue
		}
		// 添加顶点以及设置依赖关系
		flow.Add(stepFn).WithDeps(func() []*dag.Step {
			var stepFns []*dag.Step
			for _, name := range step.DependsOn {
				_stepFn, _ok := stepFnMap[name]
				if !_ok {
					continue
				}
				stepFns = append(stepFns, _stepFn)
			}
			return stepFns
		}()...)
	}

	if _, err := flow.Compile(); err != nil {
		return err
	}
	return nil
}
