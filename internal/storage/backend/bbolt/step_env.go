package bbolt

import (
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type stepEnv struct {
	db    *bbolt.DB
	tName []byte
	sName []byte
}

func (s *stepEnv) List() (res models.Envs) {
	_ = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
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

func (s *stepEnv) Create(envs ...*models.Env) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
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

func (s *stepEnv) Get(name string) (string, error) {
	var value string
	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
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

func (s *stepEnv) Delete(name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
			utils.Join(bucketPrefix, envPrefix),
		)
		if err != nil {
			return err
		}
		return bucket.Delete([]byte(name))
	})
}

func (s *stepEnv) DeleteAll() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		return bucket.DeleteBucket(utils.Join(bucketPrefix, envPrefix))
	})
}
