package sqlite

import (
	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type stepLog struct {
	*gorm.DB
	tName string
	sName string
}

func (s *stepLog) List() (res models.Logs) {
	s.Model(&tables.StepLog{}).
		Where(map[string]interface{}{
			"task_name": s.tName,
			"step_name": s.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (s *stepLog) Insert(log *models.Log) (err error) {
	return s.Create(&tables.StepLog{
		TaskName: s.tName,
		StepName: s.sName,
		Log:      *log,
	}).Error
}

func (s *stepLog) RemoveAll() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"step_name": s.sName,
	}).Delete(&tables.StepLog{}).Error
}
