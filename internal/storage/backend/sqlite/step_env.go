package sqlite

import (
	"errors"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type stepEnv struct {
	db    *gorm.DB
	tName string
	sName string
}

func (s *stepEnv) List() (res models.Envs) {
	s.db.
		Model(&tables.StepEnv{}).
		Where(map[string]interface{}{
			"task_name": s.sName,
			"step_name": s.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (s *stepEnv) Create(envs ...*models.Env) (err error) {
	if len(envs) == 0 {
		return
	}
	var _envs []tables.StepEnv
	for _, env := range envs {
		_envs = append(_envs, tables.StepEnv{
			TaskName: s.tName,
			StepName: s.sName,
			Env:      *env,
		})
	}
	return s.db.
		Create(&_envs).
		Error
}

func (s *stepEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = s.db.
		Model(&tables.StepEnv{}).
		Select("value").
		Where(map[string]interface{}{
			"task_name": s.sName,
			"step_name": s.sName,
			"name":      name,
		}).
		Scan(&res).
		Error
	return
}

func (s *stepEnv) Delete(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return s.db.
		Where(map[string]interface{}{
			"task_name": s.sName,
			"step_name": s.sName,
			"name":      name,
		}).
		Delete(&tables.StepEnv{}).
		Error
}

func (s *stepEnv) DeleteAll() (err error) {
	return s.db.
		Where(map[string]interface{}{
			"task_name": s.sName,
			"step_name": s.sName,
		}).
		Delete(&tables.StepEnv{}).
		Error
}
