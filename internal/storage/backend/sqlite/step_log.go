package sqlite

import (
	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type stepLog struct {
	db    *gorm.DB
	tName string
	sName string
}

func (s *stepLog) List() (res models.Logs) {
	s.db.
		Model(&tables.StepLog{}).
		Where(map[string]interface{}{
			"task_name": s.sName,
			"step_name": s.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (s *stepLog) Create(log *models.Log) (err error) {
	return s.db.
		Create(&tables.StepLog{
			TaskName: s.tName,
			StepName: s.sName,
			Log:      *log,
		}).
		Error
}

func (s *stepLog) DeleteAll() (err error) {
	return s.db.
		Where(map[string]interface{}{
			"task_name": s.sName,
			"step_name": s.sName,
		}).
		Delete(&tables.StepLog{}).
		Error
}
