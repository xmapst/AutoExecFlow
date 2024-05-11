package bbolt

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type task struct {
	db   *bbolt.DB
	name string
}

func (t *task) ClearAll() {
	_ = t.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(taskPrefix + t.name))
	})
}

func (t *task) Delete() (err error) {
	return t.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(taskPrefix + t.name))
	})
}

func (t *task) GetState() (state int, err error) {
	err = t.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(taskPrefix + t.name))
		if bucket == nil {
			return nil
		}
		v := bucket.Get([]byte("state"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &state)
	})
	return
}

func (t *task) Env() backend.IEnv {
	return &taskEnv{
		db:       t.db,
		taskName: t.name,
	}
}

func (t *task) Name() string {
	return t.name
}

func (t *task) Timeout() (res time.Duration, err error) {
	err = t.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(taskPrefix + t.name))
		if bucket == nil {
			return nil
		}
		v := bucket.Get([]byte("timeout"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &res)
	})
	return
}

func (t *task) Get() (res *models.Task, err error) {
	res = new(models.Task)
	err = t.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(taskPrefix + t.name))
		if bucket == nil {
			return errors.New("not found")
		}
		return utils.NewHelper(bucket).Read(res)
	})
	return res, err
}

func (t *task) Create(task *models.Task) error {
	task.Name = t.name
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + t.name))
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(task)
	})
}

func (t *task) Update(value *models.TaskUpdate) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + t.name))
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(value)
	})
}

func (t *task) Step(name string) backend.IStep {
	return &step{
		db:       t.db,
		taskName: t.name,
		name:     name,
	}
}

func (t *task) StepList(str string) (res models.Steps) {
	str = stepPrefix + str
	_ = t.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + t.name))
		if taskBucket == nil {
			return nil
		}
		return taskBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), str) {
				return nil
			}
			stepBucket := taskBucket.Bucket(k)
			if stepBucket == nil {
				return nil
			}
			var data = new(models.Step)
			err := utils.NewHelper(stepBucket).Read(data)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			res = append(res, data)
			return nil
		})
	})
	sort.Sort(res)
	return res
}

type taskEnv struct {
	db       *bbolt.DB
	taskName string
}

func (t *taskEnv) List() (res models.Envs) {
	_ = t.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + t.taskName))
		if taskBucket == nil {
			return nil
		}
		return taskBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), envPrefix) {
				return nil
			}
			envBucket := taskBucket.Bucket(k)
			if envBucket == nil {
				return nil
			}
			var data = new(models.Env)
			err := utils.NewHelper(envBucket).Read(data)
			if err != nil {
				return nil
			}
			res = append(res, data)
			return nil
		})
	})
	return res
}

func (t *taskEnv) Create(envs ...*models.Env) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		taskBucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + t.taskName))
		if err != nil {
			return err
		}
		for _, env := range envs {
			envBucket, err := taskBucket.CreateBucketIfNotExists([]byte(envPrefix + env.Name))
			if err != nil {
				return err
			}
			if err = utils.NewHelper(envBucket).Write(env); err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *taskEnv) Get(name string) (string, error) {
	var value string
	err := t.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + t.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		envBucket := taskBucket.Bucket([]byte(envPrefix + name))
		if envBucket == nil {
			return errors.New("not found")
		}
		var data = new(models.Env)
		err := utils.NewHelper(envBucket).Read(data)
		if err != nil {
			return err
		}
		value = data.Value
		return nil
	})
	return value, err
}

func (t *taskEnv) Delete(name string) (err error) {
	return t.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + t.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		return taskBucket.DeleteBucket([]byte(envPrefix + name))
	})
}

func (t *taskEnv) DeleteAll() (err error) {
	return t.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + t.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		return taskBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), envPrefix) {
				return nil
			}
			return taskBucket.DeleteBucket(k)
		})
	})
}
