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
	s.db.Model(&tables.Log{}).Where("task_name = ? AND step_name = ?", s.tName, s.sName).Order("id ASC").Find(&res)
	return
}

func (s *stepLog) Create(log *models.Log) (err error) {
	return s.db.Create(&tables.Log{
		TaskName: s.tName,
		StepName: s.sName,
		Log:      *log,
	}).Error
}
