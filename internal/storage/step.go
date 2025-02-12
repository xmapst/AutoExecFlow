package storage

import (
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sStep struct {
	*gorm.DB
	genv  IEnv
	tName string
	sName string

	env    IEnv
	depend IDepend
	log    ILog
}

func (s *sStep) Name() string {
	return s.sName
}

func (s *sStep) ClearAll() error {
	if err := s.Remove(); err != nil {
		return err
	}
	if err := s.Env().RemoveAll(); err != nil {
		return err
	}
	if err := s.Depend().RemoveAll(); err != nil {
		return err
	}
	return s.Log().RemoveAll()
}

func (s *sStep) Remove() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"name":      s.sName,
	}).Delete(&models.SStep{}).Error
}

func (s *sStep) State() (state models.State, err error) {
	err = s.Model(&models.SStep{}).
		Select("state").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&state).
		Error
	return
}

func (s *sStep) IsDisable() (disable bool) {
	if s.Model(&models.SStep{}).
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

func (s *sStep) Env() IEnv {
	if s.env == nil {
		s.env = &sStepEnv{
			DB:    s.DB,
			tName: s.tName,
			sName: s.sName,
		}
	}
	return s.env
}

func (s *sStep) TaskName() string {
	return s.tName
}

func (s *sStep) Timeout() (res time.Duration, err error) {
	err = s.Model(&models.SStep{}).
		Select("timeout").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *sStep) Type() (res string, err error) {
	err = s.Model(&models.SStep{}).
		Select("type").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *sStep) Content() (res string, err error) {
	err = s.Model(&models.SStep{}).
		Select("content").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *sStep) Action() (res string, err error) {
	err = s.Model(&models.SStep{}).
		Select("action").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *sStep) Rule() (res string, err error) {
	err = s.Model(&models.SStep{}).
		Select("rule").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Scan(&res).
		Error
	return
}

func (s *sStep) Get() (res *models.SStep, err error) {
	res = new(models.SStep)
	err = s.Model(&models.SStep{}).
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		First(res).
		Error
	return
}

func (s *sStep) Update(value *models.SStepUpdate) (err error) {
	if value == nil {
		return
	}
	return s.Model(&models.SStep{}).
		Where(map[string]interface{}{
			"task_name": s.tName,
			"name":      s.sName,
		}).
		Updates(value).
		Error
}

func (s *sStep) GlobalEnv() IEnv {
	if s.genv == nil {
		s.genv = &sTaskEnv{
			DB:    s.DB,
			tName: s.tName,
		}
	}
	return s.genv
}

// CheckDependentModel 获取依赖当前步骤的步骤
func (s *sStep) CheckDependentModel() (res bool) {
	var count int64
	s.Debug().Table("t_step").Select("count(DISTINCT t_step.name)").
		Joins("INNER JOIN t_step_depend ON t_step.task_name = t_step_depend.task_name AND t_step.name = t_step_depend.step_name").
		Where("t_step.task_name = ? AND t_step_depend.name = ? AND t_step.action != '' AND t_step.rule != ''", s.tName, s.sName).
		Count(&count)
	return count > 0
}

func (s *sStep) Depend() IDepend {
	if s.depend == nil {
		s.depend = &sStepDepend{
			DB:    s.DB,
			tName: s.tName,
			sName: s.sName,
		}
	}
	return s.depend
}

func (s *sStep) Log() ILog {
	if s.log == nil {
		s.log = &sStepLog{
			DB:    s.DB,
			tName: s.tName,
			sName: s.sName,
		}
	}
	return s.log
}
