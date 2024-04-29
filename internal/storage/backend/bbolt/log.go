package bbolt

import (
	"fmt"
	"sort"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type log struct {
	db       *bbolt.DB
	taskName string
	stepName string
}

func (l *log) List() (res models.Logs) {
	_ = l.db.View(func(tx *bbolt.Tx) error {
		taskBucket := tx.Bucket([]byte(taskPrefix + l.taskName))
		if taskBucket == nil {
			return nil
		}
		stepBucket := taskBucket.Bucket([]byte(stepPrefix + l.stepName))
		if stepBucket == nil {
			return nil
		}
		return stepBucket.ForEach(func(k, v []byte) error {
			if !strings.HasPrefix(string(k), logPrefix) {
				return nil
			}
			logBucket := stepBucket.Bucket(k)
			if logBucket == nil {
				return nil
			}
			var data = new(models.Log)
			err := utils.NewHelper(logBucket).Read(data)
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

func (l *log) Create(log *models.Log) error {
	log.TaskName = l.taskName
	log.StepName = l.stepName
	return l.db.Update(func(tx *bbolt.Tx) error {
		taskBucket, err := tx.CreateBucketIfNotExists([]byte(taskPrefix + l.taskName))
		if err != nil {
			return err
		}
		stepBucket, err := taskBucket.CreateBucketIfNotExists([]byte(stepPrefix + l.stepName))
		if err != nil {
			return err
		}
		logBucket, err := stepBucket.CreateBucketIfNotExists([]byte(fmt.Sprintf("%s%d", logPrefix, log.Line)))
		if err != nil {
			return err
		}
		return utils.NewHelper(logBucket).Write(log)
	})
}
