package bbolt

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/avast/retry-go/v4"
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

var (
	bucketPrefix = []byte("Obj")
	taskPrefix   = []byte("Task")
	stepPrefix   = []byte("Step")
	envPrefix    = []byte("Env")
	dependPrefix = []byte("Dep")
	logPrefix    = []byte("Log")
)

type Bolt struct {
	*bbolt.DB
}

func New(path string) (backend.IStorage, error) {
	var b = new(Bolt)
	err := retry.Do(
		func() (err error) {
			b.DB, err = bbolt.Open(filepath.Join(path, "database.db"), os.ModePerm, &bbolt.Options{})
			if err != nil {
				// 尝试删除后重建
				_ = os.RemoveAll(path)
				_ = os.MkdirAll(path, os.ModeDir)
				return err
			}
			return
		},
		retry.Attempts(3),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			_max := time.Duration(n)
			if _max > 8 {
				_max = 8
			}
			duration := time.Second * _max * _max
			return duration
		}),
	)
	if err != nil {
		return nil, err
	}
	for _, t := range b.TaskList(backend.All) {
		if *t.State != models.Running {
			continue
		}
		_ = b.Task(t.Name).Update(&models.TaskUpdate{
			State:   models.Pointer(models.Failed),
			ETime:   models.Pointer(time.Now()),
			Message: "unexpected ending",
		})
		for _, s := range b.Task(t.Name).StepList(backend.All) {
			if *s.State != models.Running {
				continue
			}
			_ = b.Task(t.Name).Step(s.Name).Update(&models.StepUpdate{
				State:   models.Pointer(models.Failed),
				ETime:   models.Pointer(time.Now()),
				Code:    models.Pointer(int64(-999)),
				Message: "unexpected ending",
			})
		}
	}
	return b, nil
}

func (b *Bolt) Name() string {
	return "bbolt"
}

func (b *Bolt) Close() error {
	if err := b.Sync(); err != nil {
		return err
	}
	return b.DB.Close()
}

func (b *Bolt) Task(name string) backend.ITask {
	return &task{
		db:    b.DB,
		tName: []byte(name),
	}
}

func (b *Bolt) TaskList(str string) (res models.Tasks) {
	strPrefix := utils.Join(bucketPrefix, taskPrefix, []byte(str))
	_ = b.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			if !bytes.HasPrefix(name, strPrefix) {
				return nil
			}
			var data = &models.Task{}
			err := utils.NewHelper(b).Read(data)
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
