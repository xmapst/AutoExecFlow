package sqlite

import (
	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type log struct {
	db       *gorm.DB
	taskName string
	stepName string
}

func (l *log) List() (res []*models.Log) {
	l.db.Model(&tables.Log{}).Where("task_name = ? AND step_name = ?", l.taskName, l.stepName).Order("id ASC").Find(&res)
	return
}

func (l *log) Create(log *models.Log) (err error) {
	log.TaskName = l.taskName
	log.StepName = l.stepName
	return l.db.Create(&tables.Log{
		Log: *log,
	}).Error
}
