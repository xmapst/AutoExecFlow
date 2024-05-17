package sqlite

import (
	"errors"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type taskEnv struct {
	db    *gorm.DB
	tName string
}

func (t *taskEnv) List() (res models.Envs) {
	t.db.Model(&tables.TaskEnv{}).
		Where("task_name = ?", t.tName).
		Order("id ASC").
		Find(&res)
	return
}

func (t *taskEnv) Create(envs ...*models.Env) (err error) {
	if len(envs) == 0 {
		return
	}
	var _envs []tables.TaskEnv
	for _, value := range envs {
		_envs = append(_envs, tables.TaskEnv{
			TaskName: t.tName,
			Env:      *value,
		})
	}
	return t.db.Create(&_envs).Error
}

func (t *taskEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = t.db.Model(&tables.TaskEnv{}).
		Select("value").
		Where("task_name = ? AND name = ?", t.tName, name).
		First(res).
		Error
	return
}

func (t *taskEnv) Delete(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return t.db.Where("task_name = ? AND name = ?", t.tName, name).
		Delete(&tables.TaskEnv{}).
		Error
}

func (t *taskEnv) DeleteAll() (err error) {
	return t.db.Where("task_name = ?", t.tName).
		Delete(&tables.TaskEnv{}).
		Error
}
