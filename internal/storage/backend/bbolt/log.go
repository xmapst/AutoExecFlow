package bbolt

import (
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/models"
)

type log struct {
	db       *bbolt.DB
	taskName string
	stepName string
}

func (l *log) List() (res []*models.Log) {
	//TODO implement me
	panic("implement me")
}

func (l *log) Create(log *models.Log) (err error) {
	//TODO implement me
	panic("implement me")
}
