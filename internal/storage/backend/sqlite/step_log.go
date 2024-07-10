package sqlite

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type stepLog struct {
	*gorm.DB
	tName string
	sName string

	lock sync.Mutex
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

func (s *stepLog) Insert(log *models.Log) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := s.Model(&tables.StepLog{}).
			Where(map[string]interface{}{
				"task_name": s.tName,
				"step_name": s.sName,
			}).
			Count(&count).Error; err != nil {
			return err
		}
		log.Line = models.Pointer(count)
		return tx.Create(&tables.StepLog{
			TaskName: s.tName,
			StepName: s.sName,
			Log:      *log,
		}).Error
	})
}

func (s *stepLog) Write(content string) {
	if err := s.Insert(&models.Log{
		Timestamp: time.Now().UnixNano(),
		Content:   content,
	}); err != nil {
		logx.Warnln(err)
	}
}

func (s *stepLog) Writef(format string, args ...interface{}) {
	s.Write(fmt.Sprintf(format, args...))
}

func (s *stepLog) RemoveAll() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"step_name": s.sName,
	}).Delete(&tables.StepLog{}).Error
}
