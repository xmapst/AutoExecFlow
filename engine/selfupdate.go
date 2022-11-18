package engine

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/config"
	"github.com/xmapst/osreapi/utils"
	"os"
	"time"
)

var TmpData selfUpdate

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
	err = json.Unmarshal(bs, &TmpData)
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
	cache.SetTask(TmpData.TaskId, taskState, taskState.Times.TTL)
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
	cache.SetTaskStep(
		fmt.Sprintf("%s:selfupdate", TmpData.TaskId),
		taskStepState, taskStepState.Times.TTL,
	)
}
