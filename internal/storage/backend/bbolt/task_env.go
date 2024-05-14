package bbolt

import (
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type taskEnv struct {
	db    *bbolt.DB
	tName []byte
}

func (t *taskEnv) List() (res models.Envs) {
	_ = t.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, t.tName),
			utils.Join(bucketPrefix, envPrefix),
		)
		if err != nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			res = append(res, &models.Env{
				Name:  string(k),
				Value: string(v),
			})
			return nil
		})
	})
	return res
}

func (t *taskEnv) Create(envs ...*models.Env) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(
			tx,
			utils.Join(bucketPrefix, taskPrefix, t.tName),
			utils.Join(bucketPrefix, envPrefix),
		)
		if err != nil {
			return err
		}
		for _, env := range envs {
			err = bucket.Put([]byte(env.Name), []byte(env.Value))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *taskEnv) Get(name string) (string, error) {
	var value string
	err := t.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, t.tName),
			utils.Join(bucketPrefix, envPrefix),
		)
		if err != nil {
			return err
		}
		value = string(bucket.Get([]byte(name)))
		return nil
	})
	return value, err
}

func (t *taskEnv) Delete(name string) error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, t.tName),
			utils.Join(bucketPrefix, envPrefix),
		)
		if err != nil {
			return err
		}
		return bucket.Delete([]byte(name))
	})
}

func (t *taskEnv) DeleteAll() error {
	return t.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(tx, utils.Join(bucketPrefix, taskPrefix, t.tName))
		if err != nil {
			return err
		}
		return bucket.DeleteBucket(utils.Join(bucketPrefix, envPrefix))
	})
}
