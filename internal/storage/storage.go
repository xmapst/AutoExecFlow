package storage

import (
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bolt"
	"github.com/xmapst/osreapi/internal/storage/types"
	"github.com/xmapst/osreapi/pkg/logx"
)

var db backend.ITaskCache

func New(t, d string) (err error) {
	switch t {
	default:
		// bolt
		db, err = bolt.New(d)
	}
	if err != nil {
		logx.Errorln(err)
		return err
	}
	return
}

func TaskList() (res types.TaskStates, err error) {
	return db.TaskList("")
}

func TaskDetail(task string) (res *types.TaskState, err error) {
	return db.TaskDetail(task)
}

func TaskStepList(task string) (res types.TaskStepStates, err error) {
	return db.TaskStepList(task)
}

func TaskStepDetail(task, step string) (res *types.TaskStepState, err error) {
	return db.TaskStepDetail(task, step)
}

func TaskStepLogList(task, step string) (res types.TaskStepLogs, err error) {
	return db.TaskStepLogList(task, step)
}

func SetTask(task string, val *types.TaskState) error {
	return db.SetTask(task, val)
}

func SetTaskStep(task, step string, val *types.TaskStepState) error {
	return db.SetTaskStep(task, step, val)
}

func SetTaskStepLog(task, step string, line int64, val *types.TaskStepLog) error {
	return db.SetTaskStepLog(task, step, line, val)
}
func Close() error {
	return db.Close()
}
