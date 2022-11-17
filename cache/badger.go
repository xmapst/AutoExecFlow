package cache

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

const (
	// Stop 0, Running 1, Pending 2
	Stop = iota
	Running
	Pending

	TaskPrefix  = "task"
	StepPrefix  = "step"
	SystemError = -255
)

var StateENMap = map[int]string{
	SystemError: "System Error",
	Stop:        "Stop",
	Running:     "Running",
	Pending:     "Pending",
}

var StateCNMap = map[int]string{
	SystemError: "系统错误",
	Stop:        "已结束",
	Running:     "执行中",
	Pending:     "等待执行",
}

type Task struct {
	Name           string
	CommandType    string
	CommandContent string
	EnvVars        []string
	DependsOn      []string
	Timeout        time.Duration
}

type TaskState struct {
	State        int    `json:"state"`
	StepCount    int    `json:"step_count"`
	HardWareID   string `json:"hard_ware_id"`
	VMInstanceID string `json:"vm_instance_id"`
	Times        *Times `json:"times,omitempty"`
}

type TaskStepState struct {
	Step      int      `json:"step"`
	Name      string   `json:"name"`
	State     int      `json:"state"`
	Code      int      `json:"code"`
	Message   string   `json:"message"`
	DependsOn []string `json:"depends_on"`
	Times     *Times   `json:"times"`
}

type Times struct {
	Begin int64         `json:"begin,omitempty"`
	End   int64         `json:"end,omitempty"`
	TTL   time.Duration `json:"ttl,omitempty"`
}

var (
	db   *badger.DB
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func Init() {
	var err error
	opt := badger.DefaultOptions("").
		WithInMemory(true).
		WithLogger(logrus.StandardLogger()).
		WithLoggingLevel(badger.INFO)
	db, err = badger.Open(opt)
	if err != nil {
		logrus.Fatalln(err)
	}
}

func Close() {
	_ = db.Sync()
	err := db.Close()
	if err != nil {
		logrus.Error(err)
	}
}

func GetAllTaskState() (res map[string]TaskState, err error) {
	prefix := fmt.Sprintf("%s:", TaskPrefix)
	res = make(map[string]TaskState)
	_ = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := strings.TrimPrefix(string(item.Key()), prefix)
			err = item.Value(func(v []byte) error {
				var taskState = new(TaskState)
				err = json.Unmarshal(v, taskState)
				if err != nil {
					logrus.Errorln(err)
					return err
				}
				ttl := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
				taskState.Times.TTL = ttl
				res[key] = *taskState
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

type ListData struct {
	ID    string `json:"id,omitempty"`
	State int    `json:"state,omitempty"`
	Count int    `json:"step_count,omitempty"`
	Times *Times `json:"times,omitempty"`
}

type ListByStartTimes []*ListData

func (l ListByStartTimes) Len() int           { return len(l) }
func (l ListByStartTimes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ListByStartTimes) Less(i, j int) bool { return l[i].Times.Begin < l[j].Times.Begin }

func GetAllByBeginTime() (res ListByStartTimes) {
	_res, err := GetAllTaskState()
	if err != nil {
		logrus.Error(err)
		return nil
	}
	for k, v := range _res {
		res = append(res, &ListData{
			ID:    k,
			State: v.State,
			Times: v.Times,
			Count: v.StepCount,
		})
	}
	// sort by StartTimes
	sort.Sort(res)
	return
}

type ListByCompletedTimes []*ListData

func (l ListByCompletedTimes) Len() int           { return len(l) }
func (l ListByCompletedTimes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ListByCompletedTimes) Less(i, j int) bool { return l[i].Times.End < l[j].Times.End }

func GetAllByEndTime() (res ListByCompletedTimes) {
	_res, err := GetAllTaskState()
	if err != nil {
		logrus.Error(err)
		return nil
	}
	for k, v := range _res {
		res = append(res, &ListData{
			ID:    k,
			State: v.State,
			Times: v.Times,
			Count: v.StepCount,
		})
	}
	// sort by CompletedTimes
	sort.Sort(res)
	return
}

type ListByExpiredTimes []*ListData

func (l ListByExpiredTimes) Len() int           { return len(l) }
func (l ListByExpiredTimes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ListByExpiredTimes) Less(i, j int) bool { return l[i].Times.TTL < l[j].Times.TTL }

func GetAllByTTLTime() (res ListByExpiredTimes) {
	_res, err := GetAllTaskState()
	if err != nil {
		logrus.Error(err)
		return nil
	}
	for k, v := range _res {
		res = append(res, &ListData{
			ID:    k,
			State: v.State,
			Times: v.Times,
			Count: v.StepCount,
		})
	}
	// sort by ExpiredTimes
	sort.Sort(res)
	return
}

type TaskStepStates []*TaskStepState

func (e TaskStepStates) Len() int           { return len(e) }
func (e TaskStepStates) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e TaskStepStates) Less(i, j int) bool { return e[i].Step < e[j].Step }

func GetTaskState(id string) (value *TaskState, found bool) {
	key := fmt.Sprintf("%s:%s", TaskPrefix, id)
	var err error
	var item *badger.Item
	err = db.View(func(txn *badger.Txn) error {
		item, err = txn.Get([]byte(key))
		return err
	})

	if err != nil {
		if err != badger.ErrKeyNotFound {
			logrus.Error(err)
		}
		return nil, false
	}
	var _value = new(TaskState)
	var val []byte
	val, err = getItemValue(item)
	err = json.Unmarshal(val, _value)
	if err != nil {
		logrus.Error(err)
		return nil, false
	}
	ttl := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
	_value.Times.TTL = ttl
	return _value, true
}

func getItemValue(item *badger.Item) (val []byte, err error) {
	var v []byte
	err = item.Value(func(val []byte) error {
		v = append(v, val...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return v, err
}

func Del(key string) {
	bsk := []byte(key)
	err := db.Update(func(txn *badger.Txn) error {
		return txn.Delete(bsk)
	})
	if err != nil {
		logrus.Error(err)
	}
}

func Set(key string, val interface{}, ttl time.Duration) {
	bsv, err := json.Marshal(val)
	if err != nil {
		logrus.Error(err)
		return
	}
	bsk := []byte(key)
	err = db.Update(func(txn *badger.Txn) error {
		if ttl != 0 {
			return txn.SetEntry(badger.NewEntry(bsk, bsv).WithTTL(ttl))
		}
		return txn.Set(bsk, bsv)
	})
	if err != nil {
		logrus.Error(err)
	}
}

func SetTask(key string, val interface{}, ttl time.Duration) {
	key = fmt.Sprintf("%s:%s", TaskPrefix, key)
	Set(key, val, ttl)
}

func SetTaskStep(key string, val interface{}, ttl time.Duration) {
	key = fmt.Sprintf("%s:%s", StepPrefix, key)
	Set(key, val, ttl)
}

func GetAllTaskStepState(taskID string) TaskStepStates {
	prefix := fmt.Sprintf("%s:%s:", StepPrefix, taskID)
	var taskStepStates TaskStepStates
	_ = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var taskStepState = new(TaskStepState)
				err := json.Unmarshal(v, taskStepState)
				if err != nil {
					logrus.Errorln(err)
					return err
				}
				ttl := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
				taskStepState.Times.TTL = ttl
				taskStepStates = append(taskStepStates, taskStepState)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	sort.Sort(taskStepStates)
	return taskStepStates
}