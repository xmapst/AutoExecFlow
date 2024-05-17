package sqlite

import (
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type step struct {
	db    *gorm.DB
	tName string
	sName string
}

func (s *step) Name() string {
	return s.sName
}

func (s *step) ClearAll() {
	_ = s.Delete()
	_ = s.Env().DeleteAll()
	_ = s.Depend().DeleteAll()
	_ = s.Log().DeleteAll()
}

func (s *step) Delete() (err error) {
	return s.db.Where("task_name = ? AND name = ?", s.tName, s.sName).
		Delete(&tables.Step{}).
		Error
}

func (s *step) State() (state int, err error) {
	var data = new(tables.Step)
	err = s.db.Model(&tables.Step{}).
		Select("state").
		Where("task_name = ? AND name = ?", s.tName, s.sName).
		First(&data).
		Error
	state = *data.State
	return
}

func (s *step) Env() backend.IEnv {
	return &stepEnv{
		db:    s.db,
		tName: s.tName,
		sName: s.sName,
	}
}

func (s *step) TaskName() string {
	return s.tName
}

func (s *step) Timeout() (res time.Duration, err error) {
	var data = new(tables.Step)
	err = s.db.Model(&tables.Step{}).
		Select("timeout").
		Where("task_name = ? AND name = ?", s.tName, s.sName).
		First(&data).
		Error
	res = data.Timeout
	return
}

func (s *step) Type() (res string, err error) {
	var data = new(tables.Step)
	err = s.db.Model(&tables.Step{}).
		Select("type").
		Where("task_name = ? AND name = ?", s.tName, s.sName).
		First(&data).
		Error
	res = data.Type
	return
}

func (s *step) Content() (res string, err error) {
	var data = new(tables.Step)
	err = s.db.Model(&tables.Step{}).
		Select("content").
		Where("task_name = ? AND name = ?", s.tName, s.sName).
		First(&data).
		Error
	res = data.Content
	return
}

func (s *step) Get() (res *models.Step, err error) {
	res = new(models.Step)
	err = s.db.Model(&tables.Step{}).
		Where("task_name = ? AND name = ?", s.tName, s.sName).
		First(res).
		Error
	return
}

func (s *step) Create(step *models.Step) (err error) {
	step.Name = s.sName
	return s.db.Create(&tables.Step{
		TaskName: s.tName,
		Step:     *step,
	}).Error
}

func (s *step) Update(value *models.StepUpdate) (err error) {
	if value == nil {
		return
	}
	return s.db.Model(&tables.Step{}).
		Where("task_name = ? AND name = ?", s.tName, s.sName).
		Updates(value).
		Error
}

func (s *step) Depend() backend.IDepend {
	return &stepDepend{
		db:    s.db,
		tName: s.tName,
		sName: s.sName,
	}
}

func (s *step) Log() backend.ILog {
	return &stepLog{
		db:    s.db,
		tName: s.tName,
		sName: s.sName,
	}
}
