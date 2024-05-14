package bbolt

import (
	"encoding/json"
	"time"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type step struct {
	db    *bbolt.DB
	tName []byte
	sName []byte
}

func (s *step) Name() string {
	return string(s.sName)
}

func (s *step) ClearAll() {
	_ = s.Delete()
}

func (s *step) Delete() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(tx, utils.Join(bucketPrefix, taskPrefix, s.tName))
		if err != nil {
			return err
		}
		return bucket.DeleteBucket(utils.Join(bucketPrefix, stepPrefix, s.sName))
	})
}

func (s *step) State() (state int, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		v := bucket.Get([]byte("state"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &state)
	})
	return state, err
}

func (s *step) Env() backend.IEnv {
	return &stepEnv{
		db:    s.db,
		tName: s.tName,
		sName: s.sName,
	}
}

func (s *step) TaskName() string {
	return string(s.tName)
}

func (s *step) Timeout() (res time.Duration, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
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

func (s *step) Type() (res string, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		v := bucket.Get([]byte("type"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &res)
	})
	return
}

func (s *step) Content() (res string, err error) {
	err = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		v := bucket.Get([]byte("content"))
		if v == nil {
			return nil
		}
		return json.Unmarshal(v, &res)
	})
	return
}

func (s *step) Get() (res *models.Step, err error) {
	res = new(models.Step)
	err = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Read(res)
	})
	return res, err
}

func (s *step) Create(step *models.Step) error {
	step.Name = string(s.sName)
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(step)
	})
}

func (s *step) Update(value *models.StepUpdate) (err error) {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(value)
	})
}

func (s *step) Depend() backend.IDepend {
	return &stepDepend{
		db:    s.db,
		tName: s.tName,
		sName: s.sName,
	}
}

func (s *step) Log() backend.ILog {
	return &stepLog{
		db:    s.db,
		tName: s.tName,
		sName: s.sName,
	}
}
