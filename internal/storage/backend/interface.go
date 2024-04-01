package backend

import (
	"errors"
	"time"

	"github.com/xmapst/osreapi/internal/storage/types"
)

type ITaskCache interface {
	TaskList(prefix string) (res types.TaskStates, err error)
	TaskDetail(task string) (res *types.TaskState, err error)
	TaskStepList(task string) (res types.TaskStepStates, err error)
	TaskStepDetail(task, step string) (res *types.TaskStepState, err error)
	TaskStepLogList(task, step string) (res types.TaskStepLogs, err error)
	SetTask(task string, val *types.TaskState) error
	SetTaskStep(task, step string, val *types.TaskStepState) error
	SetTaskStepLog(task, step string, line int64, val *types.TaskStepLog) error
	Close() (err error)
	Name() string
}

type Value struct {
	TTL   time.Duration
	Value []byte
}

func SafeCopy(des, src []byte) []byte {
	if len(des) < len(src) {
		des = make([]byte, len(src))
	} else {
		des = des[:len(src)]
	}
	copy(des, src)
	return des
}

var ErrNotExist = errors.New("does not exist")
