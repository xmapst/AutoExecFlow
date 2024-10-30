package storage

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sTaskEnv struct {
	*gorm.DB
	tName string
}

func (t *sTaskEnv) List() (res models.SEnvs) {
	t.Model(&models.STaskEnv{}).
		Where(map[string]interface{}{
			"task_name": t.tName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (t *sTaskEnv) Insert(envs ...*models.SEnv) (err error) {
	if len(envs) == 0 {
		return
	}
	var _envs []models.STaskEnv
	for _, env := range envs {
		_envs = append(_envs, models.STaskEnv{
			TaskName: t.tName,
			SEnv:     *env,
		})
	}
	return t.Create(&_envs).Error
}

func (t *sTaskEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = t.Model(&models.STaskEnv{}).
		Select("value").
		Where(map[string]interface{}{
			"task_name": t.tName,
			"name":      name,
		}).
		Scan(&res).
		Error
	return
}

func (t *sTaskEnv) Remove(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return t.Where(map[string]interface{}{
		"task_name": t.tName,
		"name":      name,
	}).Delete(&models.STaskEnv{}).Error
}

func (t *sTaskEnv) RemoveAll() (err error) {
	return t.Where(map[string]interface{}{
		"task_name": t.tName,
	}).Delete(&models.STaskEnv{}).Error
}
