package storage

import (
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/logx"
)

var db backend.IStorage

func New(t, d string) (err error) {
	switch t {
	default:
		db, err = sqlite.New(d)
	}
	if err != nil {
		logx.Errorln(err)
		return err
	}
	return
}

func Name() string {
	return db.Name()
}

func Close() error {
	return db.Close()
}

func Task(name string) backend.ITask {
	return db.Task(name)
}

func TaskCount() (res int64) {
	return db.TaskCount()
}

func TaskList(page, pageSize int64, str string) (res []*models.Task, total int64) {
	return db.TaskList(page, pageSize, str)
}