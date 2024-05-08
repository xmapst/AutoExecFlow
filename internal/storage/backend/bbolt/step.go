package bbolt

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type step struct {
	db       *bbolt.DB
	taskName string
	name     string
}

func (s *step) ClearAll() {
	_ = s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return nil
		}
		return taskBucket.DeleteBucket([]byte(stepPrefix + s.name))
	})
}

func (s *step) Name() string {
	return s.name
}

func (s *step) TaskName() string {
	return s.taskName
}

func (s *step) Timeout() (res time.Duration, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.name))
		if stepBucket == nil {
			return errors.New("not found")
		}
		v := stepBucket.Get([]byte("timeout"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &res)
	})
	return
}

func (s *step) Type() (res string, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.name))
		if stepBucket == nil {
			return errors.New("not found")
		}
		v := stepBucket.Get([]byte("type"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &res)
	})
	return
}

func (s *step) Content() (res string, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.name))
		if stepBucket == nil {
			return errors.New("not found")
		}
		v := stepBucket.Get([]byte("content"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &res)
	})
	return
}

func (s *step) Delete() (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return nil
		}
		return taskBucket.DeleteBucket([]byte(stepPrefix + s.name))
	})
}

func (s *step) GetState() (state int, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.name))
		if stepBucket == nil {
			return errors.New("not found")
		}
		v := stepBucket.Get([]byte("state"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &state)
	})
	return state, err
}

func (s *step) Env() backend.IEnv {
	return &stepEnv{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

func (s *step) Get() (res *models.Step, err error) {
	res = new(models.Step)
	err = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.name))
		if stepBucket == nil {
			return errors.New("not found")
		}
		return utils.NewHelper(stepBucket).Read(res)
	})
	return res, err
}

func (s *step) Create(step *models.Step) error {
	step.TaskName = s.taskName
	step.Name = s.name
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + s.taskName))
		if err != nil {
			return err
		}
		stepBucket, err := taskBucket.CreateBucketIfNotExists([]byte(stepPrefix + s.name))
		if err != nil {
			return err
		}
		return utils.NewHelper(stepBucket).Write(step)
	})
}

func (s *step) Update(value *models.StepUpdate) (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + s.taskName))
		if err != nil {
			return err
		}
		stepBucket, err := taskBucket.CreateBucketIfNotExists([]byte(stepPrefix + s.name))
		if err != nil {
			return err
		}
		return utils.NewHelper(stepBucket).Write(value)
	})
}

func (s *step) Depend() backend.IDepend {
	return &stepDepend{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

func (s *step) Log() backend.ILog {
	return &log{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

type stepEnv struct {
	db       *bbolt.DB
	taskName string
	stepName string
}

func (s *stepEnv) List() (res models.Envs) {
	_ = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return nil
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.stepName))
		if stepBucket == nil {
			return nil
		}
		return stepBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), envPrefix) {
				return nil
			}
			envBucket := stepBucket.Bucket(k)
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

func (s *stepEnv) Create(envs ...*models.Env) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + s.taskName))
		if err != nil {
			return err
		}
		stepBucket, err := taskBucket.CreateBucketIfNotExists([]byte(stepPrefix + s.stepName))
		if err != nil {
			return err
		}
		for _, env := range envs {
			envBucket, err := stepBucket.CreateBucketIfNotExists([]byte(envPrefix + env.Name))
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

func (s *stepEnv) Get(name string) (string, error) {
	var value string
	err := s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.stepName))
		if stepBucket == nil {
			return errors.New("not found")
		}
		envBucket := stepBucket.Bucket([]byte(envPrefix + name))
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

func (s *stepEnv) Delete(name string) (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.stepName))
		if stepBucket == nil {
			return errors.New("not found")
		}
		return stepBucket.DeleteBucket([]byte(envPrefix + name))
	})
}

func (s *stepEnv) DeleteAll() (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.stepName))
		if stepBucket == nil {
			return errors.New("not found")
		}
		return stepBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), envPrefix) {
				return nil
			}
			return stepBucket.DeleteBucket(k)
		})
	})
}

type stepDepend struct {
	db       *bbolt.DB
	taskName string
	stepName string
}

func (s *stepDepend) List() (res []string) {
	_ = s.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return nil
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.stepName))
		if stepBucket == nil {
			return nil
		}
		return stepBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), dependPrefix) {
				return nil
			}
			depBucket := stepBucket.Bucket(k)
			if depBucket == nil {
				return nil
			}
			value := depBucket.Get([]byte("name"))
			if value == nil {
				return nil
			}
			res = append(res, string(value))
			return nil
		})
	})
	return res
}

func (s *stepDepend) Create(depends ...string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + s.taskName))
		if err != nil {
			return err
		}
		stepBucket, err := taskBucket.CreateBucketIfNotExists([]byte(stepPrefix + s.stepName))
		if err != nil {
			return err
		}
		for _, depend := range depends {
			dependBucket, err := stepBucket.CreateBucketIfNotExists([]byte(dependPrefix + depend))
			if err != nil {
				return err
			}
			err = dependBucket.Put([]byte("name"), []byte(depend))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *stepDepend) DeleteAll() (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + s.taskName))
		if taskBucket == nil {
			return errors.New("not found")
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + s.stepName))
		if stepBucket == nil {
			return errors.New("not found")
		}
		return stepBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), dependPrefix) {
				return nil
			}
			return stepBucket.DeleteBucket(k)
		})
	})
}
