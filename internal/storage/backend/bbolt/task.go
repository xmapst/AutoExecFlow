package bbolt

import (
	"bytes"
	"encoding/json"
	"sort"
	"time"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type task struct {
	db    *bbolt.DB
	tName []byte
}

func (t *task) Name() string {
	return string(t.tName)
}

func (t *task) ClearAll() {
	_ = t.Delete()
}

func (t *task) Delete() (err error) {
	return t.db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(utils.Join(bucketPrefix, taskPrefix, t.tName))
	})
}

func (t *task) State() (state int, err error) {
	err = t.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
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
		db:    t.db,
		tName: t.tName,
	}
}

func (t *task) Timeout() (res time.Duration, err error) {
	err = t.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
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
		bucket, err := utils.Bucket(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Read(res)
	})
	return res, err
}

func (t *task) Create(task *models.Task) error {
	task.Name = string(t.tName)
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(task)
	})
}

func (t *task) Update(value *models.TaskUpdate) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(value)
	})
}

func (t *task) Step(name string) backend.IStep {
	return &step{
		db:    t.db,
		tName: t.tName,
		sName: []byte(name),
	}
}

func (t *task) StepList(str string) (res models.Steps) {
	strPrefix := utils.Join(bucketPrefix, stepPrefix, []byte(str))
	_ = t.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
		}
		return bucket.ForEachBucket(func(k []byte) error {
			if !bytes.HasPrefix(k, strPrefix) {
				return nil
			}
			stepBucket := bucket.Bucket(k)
			if stepBucket == nil {
				return nil
			}
			var data = new(models.Step)
			err := utils.NewHelper(stepBucket).Read(data)
			if err != nil {
				return nil
			}
			res = append(res, data)
			return nil
		})
	})
	sort.Sort(res)
	return res
}
