package engine

import (
	"os"
	"time"

	"github.com/json-iterator/go"

	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/exec"
	"github.com/xmapst/osreapi/internal/logx"
	"github.com/xmapst/osreapi/internal/utils"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type selfUpdate struct {
	TaskId string `json:"SelfUpdateTaskId"`
}

func loadSelfUpdateData() {
	if !utils.FileOrPathExist(config.App.SelfUpdateData) {
		return
	}
	bs, err := os.ReadFile(config.App.SelfUpdateData)
	if err != nil {
		logx.Fatalln(err)
	}
	var self selfUpdate
	err = json.Unmarshal(bs, &self)
	if err != nil {
		logx.Errorln(err)
		return
	}
	taskState := &cache.TaskState{
		State: exec.Stop,
		Count: 1,
		Times: &cache.Times{
			ST: time.Now().UnixNano(),
			ET: time.Now().UnixNano(),
			RT: config.App.KeyExpire,
		},
	}
	cache.SetTask(self.TaskId, taskState, taskState.Times.RT)
	taskStepState := &cache.TaskStepState{
		ID:      0,
		Name:    "selfupdate",
		State:   exec.Stop,
		Code:    0,
		Message: config.App.ServiceName + " update completed",
		Times: &cache.Times{
			ST: time.Now().UnixNano(),
			ET: time.Now().UnixNano(),
			RT: config.App.KeyExpire,
		},
	}
	cache.SetTaskStep(self.TaskId, taskStepState.ID, taskStepState, taskStepState.Times.RT)
}
