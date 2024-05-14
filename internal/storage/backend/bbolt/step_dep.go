package bbolt

import (
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
)

type stepDepend struct {
	db    *bbolt.DB
	tName []byte
	sName []byte
}

func (s *stepDepend) List() (res []string) {
	_ = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
			utils.Join(bucketPrefix, dependPrefix),
		)
		if err != nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			res = append(res, string(k))
			return nil
		})
	})
	return res
}

func (s *stepDepend) Create(depends ...string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
			utils.Join(bucketPrefix, dependPrefix),
		)

		if err != nil {
			return err
		}
		for _, depend := range depends {
			err = bucket.Put([]byte(depend), nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *stepDepend) DeleteAll() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return err
		}
		return bucket.DeleteBucket(utils.Join(bucketPrefix, dependPrefix))
	})
}
