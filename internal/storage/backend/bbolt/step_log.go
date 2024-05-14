package bbolt

import (
	"bytes"
	"sort"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type stepLog struct {
	db    *bbolt.DB
	tName []byte
	sName []byte
}

func (s *stepLog) List() (res models.Logs) {
	_ = s.db.View(func(tx *bbolt.Tx) error {
		bucket, err := utils.Bucket(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
		)
		if err != nil {
			return nil
		}
		return bucket.ForEachBucket(func(k []byte) error {
			if !bytes.HasPrefix(k, utils.Join(bucketPrefix, logPrefix)) {
				return nil
			}
			logBucket := bucket.Bucket(k)
			if logBucket == nil {
				return nil
			}
			var data = new(models.Log)
			err = utils.NewHelper(logBucket).Read(data)
			if err != nil {
				return nil
			}
			res = append(res, data)
			return nil
		})
	})
	sort.Sort(res)
	return
}

func (s *stepLog) Create(log *models.Log) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := utils.CreateBucketIfNotExists(
			tx,
			utils.Join(bucketPrefix, taskPrefix, s.tName),
			utils.Join(bucketPrefix, stepPrefix, s.sName),
			utils.Join(bucketPrefix, logPrefix, utils.Int64ToBytes(*log.Line)),
		)
		if err != nil {
			return err
		}
		return utils.NewHelper(bucket).Write(log)
	})
}
