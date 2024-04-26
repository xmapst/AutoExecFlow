package bbolt

import (
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
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

func (b *Bolt) TaskList(str string) (res []*models.Task) {
	//TODO implement me
	panic("implement me")
}
