package sqlite

import (
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type step struct {
	*gorm.DB
	genv  backend.IEnv
	tName string
	sName string

	env    backend.IEnv
	depend backend.IDepend
	log    backend.ILog
}

func (s *step) Name() string {
	return s.sName
}

func (s *step) ClearAll() {
	_ = s.Remove()
	_ = s.Env().RemoveAll()
	_ = s.Depend().RemoveAll()
	_ = s.Log().RemoveAll()
}

func (s *step) Remove() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"name":      s.sName,
	}).Delete(&tables.Step{}).Error
}

func (s *step) State() (state models.State, err error) {
	err = s.Model(&tables.Step{}).
		Select("state").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&state).
		Error
	return
}

func (s *step) IsDisable() (disable bool) {
	if s.Model(&tables.Step{}).
		Select("disable").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&disable).
		Error != nil {
		return
	}
	return
}

func (s *step) Env() backend.IEnv {
	if s.env == nil {
		s.env = &stepEnv{
			DB:    s.DB,
			tName: s.tName,
			sName: s.sName,
		}
	}
	return s.env
}

func (s *step) TaskName() string {
	return s.tName
}

func (s *step) Timeout() (res time.Duration, err error) {
	err = s.Model(&tables.Step{}).
		Select("timeout").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *step) Type() (res string, err error) {
	err = s.Model(&tables.Step{}).
		Select("type").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *step) Content() (res string, err error) {
	err = s.Model(&tables.Step{}).
		Select("content").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *step) Get() (res *models.Step, err error) {
	res = new(models.Step)
	err = s.Model(&tables.Step{}).
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		First(res).
		Error
	return
}

func (s *step) Insert(step *models.Step) (err error) {
	step.Name = s.sName
	return s.Create(&tables.Step{
		TaskName: s.tName,
		Step:     *step,
	}).Error
}

func (s *step) Update(value *models.StepUpdate) (err error) {
	if value == nil {
		return
	}
	return s.Model(&tables.Step{}).
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Updates(value).
		Error
}

func (s *step) GlobalEnv() backend.IEnv {
	if s.genv == nil {
		s.genv = &taskEnv{
			DB:    s.DB,
			tName: s.tName,
		}
	}
	return s.genv
}

func (s *step) Depend() backend.IDepend {
	if s.depend == nil {
		s.depend = &stepDepend{
			DB:    s.DB,
			tName: s.tName,
			sName: s.sName,
		}
	}
	return s.depend
}

func (s *step) Log() backend.ILog {
	if s.log == nil {
		s.log = &stepLog{
			DB:    s.DB,
			tName: s.tName,
			sName: s.sName,
		}
	}
	return s.log
}
