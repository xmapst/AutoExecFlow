package sqlite

import (
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type task struct {
	db    *gorm.DB
	tName string
}

func (t *task) Name() string {
	return t.tName
}

func (t *task) ClearAll() {
	_ = t.Delete()
	_ = t.Env().DeleteAll()
	list := t.StepList("")
	for _, v := range list {
		t.Step(v.Name).ClearAll()
	}
}

func (t *task) Delete() (err error) {
	return t.db.Where("name = ?", t.tName).
		Delete(&tables.Task{}).
		Error
}

func (t *task) State() (state int, err error) {
	var data = new(models.Task)
	err = t.db.Model(&tables.Task{}).
		Select("state").
		Where("name = ?", t.tName).
		First(data).
		Error
	state = *data.State
	return
}

func (t *task) Env() backend.IEnv {
	return &taskEnv{
		db:    t.db,
		tName: t.tName,
	}
}

func (t *task) Timeout() (res time.Duration, err error) {
	var data = new(models.Task)
	err = t.db.Model(&tables.Task{}).
		Select("state").
		Where("name = ?", t.tName).
		First(data).
		Error
	res = data.Timeout
	return
}

func (t *task) Get() (res *models.Task, err error) {
	res = new(models.Task)
	err = t.db.Model(&tables.Task{}).
		Where("name = ?", t.tName).
		First(res).
		Error
	return
}

func (t *task) Create(task *models.Task) (err error) {
	task.Name = t.tName
	return t.db.Create(&tables.Task{
		Task: *task,
	}).Error
}

func (t *task) Update(value *models.TaskUpdate) (err error) {
	if value == nil {
		return
	}
	return t.db.Model(&tables.Task{}).
		Where("name = ?", t.tName).
		Updates(value).
		Error
}

func (t *task) Step(name string) backend.IStep {
	return &step{
		db:    t.db,
		tName: t.tName,
		sName: name,
	}
}

func (t *task) StepNameList(str string) (res []string) {
	query := t.db.Model(&tables.Step{}).
		Select("name").
		Order("id ASC").
		Where("task_name = ?", t.tName)
	if str != "" {
		query.Where("name LIKE ?", str+"%s")
	}
	query.Find(&res)
	return
}

func (t *task) StepList(str string) (res models.Steps) {
	query := t.db.Model(&tables.Step{}).
		Order("id ASC").
		Where("task_name = ?", t.tName)
	if str != "" {
		query.Where("name LIKE ?", str+"%s")
	}
	query.Find(&res)
	return
}
