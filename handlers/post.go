package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/config"
	"github.com/xmapst/osreapi/engine"
	"github.com/xmapst/osreapi/utils"
	"net/http"
	"time"
)

type PostStruct struct {
	CommandType    string   `json:"command_type" form:"command_type" binding:"required" description:"命令类型[必选]" example:"cmd"`
	CommandContent string   `json:"command_content" form:"command_content" binding:"required" description:"命令内容[必选]" example:"ping baidu.com"`
	Name           string   `json:"name,omitempty" form:"name,omitempty" description:"脚本名称[可选]" example:"script.ps1"`
	EnvVars        []string `json:"env_vars,omitempty" form:"env_vars,omitempty" description:"外部环境变量[可选]" example:"env1=value1,env2=value2"`
	DependsOn      []string `json:"depends_on,omitempty" form:"depends_on,omitempty" description:"自定义编排, 依赖步骤[可选]" example:""`
	TimeOut        string   `json:"time_out,omitempty" form:"time_out,omitempty" description:"超时时间[可选, 默认使用全局超时时间]" example:"3m"`
	timeout        time.Duration
}

type PostStructSlice []PostStruct

// Post
// @Summary 执行
// @description 执行命名或脚本
// @Tags Exec
// @Param ansync path bool false "异步"
// @Param scripts body []PostStruct true "scripts"
// @Success 200 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router / [post]
func Post(c *gin.Context) {
	render := Gin{Context: c}
	ansync := c.DefaultQuery("ansync", "false")
	var req PostStructSlice
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	if req == nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    http.StatusBadRequest,
			"message": utils.GetMsg(utils.CodeErrParam),
		})
		return
	}
	taskID := uuid.NewV4().String()
	req.fixName(taskID)
	if ansync == "false" {
		req.fixSync()
	}
	req.parseDuration()
	hardWareID, vmInstanceID := req.getHardWareIDAndVmInstanceID()
	c.Request.Header.Set(xTaskID, taskID)
	c.Writer.Header().Set(xTaskID, taskID)
	c.Set(xTaskID, taskID)
	var tasks []*cache.Task
	for _, v := range req {
		tasks = append(tasks, &cache.Task{
			Name:           v.Name,
			CommandType:    v.CommandType,
			CommandContent: v.CommandContent,
			EnvVars:        v.EnvVars,
			DependsOn:      v.DependsOn,
			Timeout:        v.timeout,
		})
	}
	// 加入池中异步处理
	go engine.Process(taskID, hardWareID, vmInstanceID, tasks)

	render.SetJson(map[string]interface{}{
		"id":        taskID,
		"timestamp": time.Now().UnixNano(),
		"count":     len(req),
	})
}

func (p PostStructSlice) fixName(taskID string) {
	for step, v := range p {
		if v.Name == "" {
			p[step].Name = fmt.Sprintf("%s-%d", taskID, step)
		}
	}
}

func (p PostStructSlice) fixSync() {
	for k := range p {
		if k == 0 {
			continue
		}
		p[k].DependsOn = []string{p[k-1].Name}
	}
}

func (p PostStructSlice) parseDuration() {
	for k, v := range p {
		_timeout, err := time.ParseDuration(v.TimeOut)
		if v.TimeOut == "" || err != nil {
			_timeout = config.App.ExecTimeOut
		}
		p[k].timeout = _timeout
	}
}

func (p PostStructSlice) getHardWareIDAndVmInstanceID() (hardWareID string, vmInstanceID string) {
	// check envs
	for k, v := range p {
		var _env []string
		m := utils.SliceToStrMap(v.EnvVars)
		for k, v := range m {
			_env = append(_env, fmt.Sprintf("%s=%s", k, v))
		}
		str, ok := m["HARDWARE_ID"]
		if ok && str != "" {
			hardWareID = str
		}
		str, ok = m["VM_INSTANCE_ID"]
		if ok && str != "" {
			vmInstanceID = str
		}
		p[k].EnvVars = _env
	}
	return
}
