package sqlite

import (
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type task struct {
	*gorm.DB
	tName string
}

func (t *task) Name() string {
	return t.tName
}

func (t *task) ClearAll() {
	_ = t.Remove()
	_ = t.Env().RemoveAll()
	list := t.StepList(backend.All)
	for _, v := range list {
		t.Step(v.Name).ClearAll()
	}
}

func (t *task) Remove() (err error) {
	return t.Where(map[string]interface{}{
		"name": t.tName,
	}).Delete(&tables.Task{}).Error
}

func (t *task) State() (state int, err error) {
	err = t.Model(&tables.Task{}).
		Select("state").
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Scan(&state).
		Error
	return
}

func (t *task) Env() backend.IEnv {
	return &taskEnv{
		DB:    t.DB,
		tName: t.tName,
	}
}

func (t *task) Timeout() (res time.Duration, err error) {
	err = t.Model(&tables.Task{}).
		Select("state").
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Scan(&res).
		Error
	return
}

func (t *task) Get() (res *models.Task, err error) {
	res = new(models.Task)
	err = t.Model(&tables.Task{}).
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		First(res).
		Error
	return
}

func (t *task) Insert(task *models.Task) (err error) {
	task.Name = t.tName
	return t.Create(&tables.Task{
		Task: *task,
	}).Error
}

func (t *task) Update(value *models.TaskUpdate) (err error) {
	if value == nil {
		return
	}
	return t.Model(&tables.Task{}).
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Updates(value).
		Error
}

func (t *task) Step(name string) backend.IStep {
	return &step{
		DB:    t.DB,
		tName: t.tName,
		sName: name,
	}
}

func (t *task) StepCount() (res int64) {
	t.Model(&tables.Step{}).Count(&res)
	return
}

func (t *task) StepNameList(str string) (res []string) {
	query := t.Model(&tables.Step{}).
		Select("name").
		Order("id ASC").
		Where(map[string]interface{}{
			"task_name": t.tName,
		})
	if str != "" {
		query.Where("name LIKE ?", str+"%s")
	}
	query.Find(&res)
	return
}

func (t *task) StepList(str string) (res models.Steps) {
	query := t.Model(&tables.Step{}).
		Order("id ASC").
		Where(map[string]interface{}{
			"task_name": t.tName,
		})
	if str != "" {
		query.Where("name LIKE ?", str+"%s")
	}
	query.Find(&res)
	return
}
