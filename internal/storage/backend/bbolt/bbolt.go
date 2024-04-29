package bbolt

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt/utils"
	"github.com/xmapst/osreapi/internal/storage/models"
)

const (
	bucketPrefix = "Obj#"
	taskPrefix   = bucketPrefix + "Task#"
	stepPrefix   = bucketPrefix + "Step#"
	envPrefix    = bucketPrefix + "Env#"
	dependPrefix = bucketPrefix + "Dep#"
	logPrefix    = bucketPrefix + "Log#"
)

type Bolt struct {
	*bbolt.DB
}

func New(path string) (backend.IStorage, error) {
	db, err := bbolt.Open(filepath.Join(path, "database.db"), os.ModePerm, &bbolt.Options{})
	if err != nil {
		return nil, err
	}
	b := &Bolt{
		DB: db,
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
		db:   b.DB,
		name: name,
	}
}

func (b *Bolt) TaskList(str string) (res models.Tasks) {
	str = taskPrefix + str
	_ = b.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			if !strings.HasPrefix(string(name), str) {
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
