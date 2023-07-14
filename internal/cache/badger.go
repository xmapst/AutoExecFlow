package cache

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/json-iterator/go"
	"github.com/xmapst/osreapi/internal/logx"
)

const (
	// Stop 0, Running 1, Pending 2
	Stop = iota
	Running
	Pending

	taskPrefix   = "task"
	stepPrefix   = "step"
	outputPrefix = "output"
	SystemError  = 255
)

var StateMap = map[int]string{
	SystemError: "System Error",
	Stop:        "Stop",
	Running:     "Running",
	Pending:     "Pending",
}

type MetaData struct {
	HardWareID   string `json:"hard_ware_id"`
	VMInstanceID string `json:"vm_instance_id"`
}

type Task struct {
	ID       string
	MetaData MetaData
	Steps    []*TaskStep
}

type TaskStep struct {
	ID             int64
	Name           string
	CommandType    string
	CommandContent string
	EnvVars        []string
	DependsOn      []string
	Timeout        time.Duration
}

type TaskState struct {
	State    int      `json:"state"`
	Count    int64    `json:"count"`
	MetaData MetaData `json:"meta_data"`
	Message  string   `json:"message"`
	Times    *Times   `json:"times,omitempty"`
}

type TaskStepState struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	State     int      `json:"state"`
	Code      int64    `json:"code"`
	Message   string   `json:"message"`
	DependsOn []string `json:"depends_on"`
	Times     *Times   `json:"times"`
}

type Times struct {
	ST int64         `json:"st,omitempty"` // 开始时间
	ET int64         `json:"et,omitempty"` // 结束时间
	RT time.Duration `json:"rt,omitempty"` // 剩余时间
}

type TaskStepOutput struct {
	Timestamp int64  `json:"timestamp"`
	Line      int64  `json:"line"`
	Content   string `json:"content"`
}

var (
	db   *badger.DB
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func New(path string) error {
	_ = os.RemoveAll(path)
	var err error
	opt := badger.DefaultOptions(path).
		// WithInMemory(true).
		WithLogger(logx.GetSubLogger()).
		WithLoggingLevel(badger.ERROR)
	db, err = badger.Open(opt)
	if err != nil {
		return err
	}
	return nil
}

func Close() {
	_ = db.Sync()
	err := db.Close()
	if err != nil {
		logx.Errorln(err)
	}
}

func getAllTask() (res map[string]TaskState, err error) {
	prefix := fmt.Sprintf("%s:", taskPrefix)
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
					logx.Errorln(err)
					return err
				}
				rt := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
				taskState.Times.RT = rt
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
	ID      string `json:"id,omitempty"`
	State   int    `json:"state,omitempty"`
	Count   int64  `json:"count,omitempty"`
	Message string `json:"message,omitempty"`
	Times   *Times `json:"times,omitempty"`
}

type ListByStartTimes []*ListData

func (l ListByStartTimes) Len() int           { return len(l) }
func (l ListByStartTimes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ListByStartTimes) Less(i, j int) bool { return l[i].Times.ST < l[j].Times.ST }

func GetAllByBeginTime() (res ListByStartTimes) {
	_res, err := getAllTask()
	if err != nil {
		logx.Errorln(err)
		return
	}
	for k, v := range _res {
		res = append(res, &ListData{
			ID:      k,
			State:   v.State,
			Times:   v.Times,
			Count:   v.Count,
			Message: v.Message,
		})
	}
	// sort by StartTimes
	sort.Sort(res)
	return
}

type ListByCompletedTimes []*ListData

func (l ListByCompletedTimes) Len() int           { return len(l) }
func (l ListByCompletedTimes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ListByCompletedTimes) Less(i, j int) bool { return l[i].Times.ET < l[j].Times.ET }

func GetAllByEndTime() (res ListByCompletedTimes) {
	_res, err := getAllTask()
	if err != nil {
		logx.Errorln(err)
		return
	}
	for k, v := range _res {
		res = append(res, &ListData{
			ID:      k,
			State:   v.State,
			Times:   v.Times,
			Count:   v.Count,
			Message: v.Message,
		})
	}
	// sort by CompletedTimes
	sort.Sort(res)
	return
}

type ListByExpiredTimes []*ListData

func (l ListByExpiredTimes) Len() int           { return len(l) }
func (l ListByExpiredTimes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l ListByExpiredTimes) Less(i, j int) bool { return l[i].Times.RT < l[j].Times.RT }

func GetAllByTTLTime() (res ListByExpiredTimes) {
	_res, err := getAllTask()
	if err != nil {
		logx.Errorln(err)
		return
	}
	for k, v := range _res {
		res = append(res, &ListData{
			ID:      k,
			State:   v.State,
			Times:   v.Times,
			Count:   v.Count,
			Message: v.Message,
		})
	}
	// sort by ExpiredTimes
	sort.Sort(res)
	return
}

type TaskStepStates []*TaskStepState

func (e TaskStepStates) Len() int           { return len(e) }
func (e TaskStepStates) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e TaskStepStates) Less(i, j int) bool { return e[i].ID < e[j].ID }

type TaskStepOutputs []*TaskStepOutput

func (e TaskStepOutputs) Len() int           { return len(e) }
func (e TaskStepOutputs) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e TaskStepOutputs) Less(i, j int) bool { return e[i].Line < e[j].Line }

func GetTask(task string) (value *TaskState, found bool) {
	key := fmt.Sprintf("%s:%s", taskPrefix, task)
	var err error
	var item *badger.Item
	err = db.View(func(txn *badger.Txn) error {
		item, err = txn.Get([]byte(key))
		return err
	})

	if err != nil {
		if err != badger.ErrKeyNotFound {
			logx.Errorln(err)
		}
		return
	}
	var _value = new(TaskState)
	var val []byte
	val, err = getItemValue(item)
	err = json.Unmarshal(val, _value)
	if err != nil {
		logx.Errorln(err)
		return
	}
	rt := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
	_value.Times.RT = rt
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
		logx.Errorln(err)
	}
}

func SetJson(key string, val interface{}, ttl time.Duration) {
	bsv, err := json.Marshal(val)
	if err != nil {
		logx.Errorln(err)
		return
	}
	Set([]byte(key), bsv, ttl)
}

func Set(key []byte, val []byte, ttl time.Duration) {
	err := db.Update(func(txn *badger.Txn) error {
		if ttl != 0 {
			return txn.SetEntry(badger.NewEntry(key, val).WithTTL(ttl))
		}
		return txn.Set(key, val)
	})
	if err != nil {
		logx.Errorln(err)
	}
}

func SetTask(task string, val interface{}, ttl time.Duration) {
	key := fmt.Sprintf("%s:%s", taskPrefix, task)
	SetJson(key, val, ttl)
}

func SetTaskStep(task string, step int64, val interface{}, ttl time.Duration) {
	key := fmt.Sprintf("%s:%s:%d", stepPrefix, task, step)
	SetJson(key, val, ttl)
}

func GetTaskStepStates(task string) TaskStepStates {
	prefix := fmt.Sprintf("%s:%s:", stepPrefix, task)
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
					logx.Errorln(err)
					return err
				}
				rt := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
				taskStepState.Times.RT = rt
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

func GetTaskStepState(task string, step int64) (*TaskStepState, bool) {
	key := fmt.Sprintf("%s:%s:%d", stepPrefix, task, step)
	var err error
	var item *badger.Item
	err = db.View(func(txn *badger.Txn) error {
		item, err = txn.Get([]byte(key))
		return err
	})

	if err != nil {
		if err != badger.ErrKeyNotFound {
			logx.Errorln(err)
		}
		return nil, false
	}
	var _value = new(TaskStepState)
	var val []byte
	val, err = getItemValue(item)
	err = json.Unmarshal(val, _value)
	if err != nil {
		logx.Errorln(err)
		return nil, false
	}
	rt := time.Unix(int64(item.ExpiresAt()), 0).Sub(time.Now())
	_value.Times.RT = rt
	return _value, true
}

func SetTaskStepOutput(task string, step, line int64, val interface{}, ttl time.Duration) {
	key := fmt.Sprintf("%s:%s:%d:%d", outputPrefix, task, step, line)
	SetJson(key, val, ttl)
}

func SetTaskStepOutputDone(task string, step, num int64, ttl time.Duration) {
	key := fmt.Sprintf("%s:%s:%d_done", outputPrefix, task, step)
	Set([]byte(key), []byte(strconv.FormatInt(num, 10)), ttl)
}

func GetTaskStepOutputDone(task string, step int64) int64 {
	key := fmt.Sprintf("%s:%s:%d_done", outputPrefix, task, step)
	var err error
	var item *badger.Item
	err = db.View(func(txn *badger.Txn) error {
		item, err = txn.Get([]byte(key))
		return err
	})

	if err != nil {
		logx.Errorln(err)
		return -1
	}
	var val []byte
	val, err = getItemValue(item)
	if err != nil {
		logx.Errorln(err)
		return -1
	}
	num, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		logx.Errorln(err)
		return -1
	}
	return num
}

func GetTaskStepAllOutput(task string, step int64) TaskStepOutputs {
	prefix := fmt.Sprintf("%s:%s:%d:", outputPrefix, task, step)
	var taskStepOutputs TaskStepOutputs
	_ = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			_ = item.Value(func(v []byte) error {
				var taskStepOutput = new(TaskStepOutput)
				err := json.Unmarshal(v, taskStepOutput)
				if err != nil {
					logx.Errorln(err)
					return err
				}
				taskStepOutputs = append(taskStepOutputs, taskStepOutput)
				return nil
			})
		}
		return nil
	})
	sort.Sort(taskStepOutputs)
	return taskStepOutputs
}
