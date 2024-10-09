package storage

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type stepEnv struct {
	*gorm.DB
	tName string
	sName string
}

func (s *stepEnv) List() (res models.Envs) {
	s.Model(&models.StepEnv{}).
		Where(map[string]interface{}{
			"task_name": s.tName,
			"step_name": s.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (s *stepEnv) Insert(envs ...*models.Env) (err error) {
	if len(envs) == 0 {
		return
	}
	var _envs []models.StepEnv
	for _, env := range envs {
		_envs = append(_envs, models.StepEnv{
			TaskName: s.tName,
			StepName: s.sName,
			Env:      *env,
		})
	}
	return s.Create(&_envs).Error
}

func (s *stepEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = s.Model(&models.StepEnv{}).
		Select("value").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"step_name": s.sName,
			"name":      name,
		}).
		Scan(&res).
		Error
	return
}

func (s *stepEnv) Remove(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"step_name": s.sName,
		"name":      name,
	}).Delete(&models.StepEnv{}).Error
}

func (s *stepEnv) RemoveAll() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"step_name": s.sName,
	}).Delete(&models.StepEnv{}).Error
}
