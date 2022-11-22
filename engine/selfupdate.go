package engine

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/config"
	"github.com/xmapst/osreapi/utils"
	"os"
	"time"
)

type selfUpdate struct {
	TaskId string `json:"SelfUpdateTaskId"`
}

func LoadSelfUpdateData() {
	if !utils.FileOrPathExist(config.App.SelfUpdateData) {
		return
	}
	bs, err := os.ReadFile(config.App.SelfUpdateData)
	if err != nil {
		logrus.Fatalln(err)
	}
	var self selfUpdate
	err = json.Unmarshal(bs, &self)
	if err != nil {
		logrus.Error(err)
		return
	}
	taskState := &cache.TaskState{
		State: cache.Stop,
		Count: 1,
		Times: &cache.Times{
			Begin: time.Now().UnixNano(),
			End:   time.Now().UnixNano(),
			TTL:   config.App.KeyExpire,
		},
	}
	cache.SetTask(self.TaskId, taskState, taskState.Times.TTL)
	taskStepState := &cache.TaskStepState{
		Step:    0,
		Name:    "selfupdate",
		State:   cache.Stop,
		Code:    0,
		Message: config.App.ServiceName + " update completed",
		Times: &cache.Times{
			Begin: time.Now().UnixNano(),
			End:   time.Now().UnixNano(),
			TTL:   config.App.KeyExpire,
		},
	}
	cache.SetTaskStep(self.TaskId, 0, taskStepState, taskStepState.Times.TTL)
}
