package storage

import (
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

var storage IStorage

const (
	TYPE_SQLITE = "sqlite"
	TYPE_MYSQL  = "mysql"
)

func New(rawURL string) error {
	db, err := newDB(rawURL)
	if err != nil {
		return err
	}
	storage = db
	return nil
}

func Name() string {
	return storage.Name()
}

func Close() error {
	return storage.Close()
}

func Task(name string) ITask {
	return storage.Task(name)
}

func TaskCreate(task *models.Task) (err error) {
	return storage.TaskCreate(task)
}

func TaskCount(state models.State) (res int64) {
	return storage.TaskCount(state)
}

func TaskList(page, pageSize int64, str string) (res []*models.Task, total int64) {
	return storage.TaskList(page, pageSize, str)
}
