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
	t.db.
		Model(&tables.TaskEnv{}).
		Where(map[string]interface{}{
			"task_name": t.tName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (t *taskEnv) Create(envs ...*models.Env) (err error) {
	if len(envs) == 0 {
		return
	}
	var _envs []tables.TaskEnv
	for _, env := range envs {
		_envs = append(_envs, tables.TaskEnv{
			TaskName: t.tName,
			Env:      *env,
		})
	}
	return t.db.
		Create(&_envs).
		Error
}

func (t *taskEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = t.db.
		Model(&tables.TaskEnv{}).
		Select("value").
		Where(map[string]interface{}{
			"task_name": t.tName,
			"name":      name,
		}).
		Scan(&res).
		Error
	return
}

func (t *taskEnv) Delete(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return t.db.
		Where(map[string]interface{}{
			"task_name": t.tName,
			"name":      name,
		}).
		Delete(&tables.TaskEnv{}).
		Error
}

func (t *taskEnv) DeleteAll() (err error) {
	return t.db.
		Where(map[string]interface{}{
			"task_name": t.tName,
		}).
		Delete(&tables.TaskEnv{}).
		Error
}
