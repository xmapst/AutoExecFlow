package sqlite

import (
	"errors"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type task struct {
	db   *gorm.DB
	name string
}

func (t *task) ClearAll() {
	_ = t.Delete()
	_ = t.Env().DeleteAll()
	list := t.StepList("")
	for _, v := range list {
		t.Step(v.Name).ClearAll()
	}
}

func (t *task) Get() (res *models.Task, err error) {
	res = new(models.Task)
	err = t.db.Model(&tables.Task{}).Where("name = ?", t.name).First(res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("not found")
	}
	return
}

func (t *task) Delete() (err error) {
	err = t.db.Where("name = ?", t.name).Delete(&tables.Task{}).Error
	return
}

func (t *task) Create(task *models.Task) (err error) {
	task.Name = t.name
	if err = t.db.Create(&tables.Task{
		Task: *task,
	}).Error; err != nil {
		return err
	}
	return
}

func (t *task) Update(value *models.TaskUpdate) (err error) {
	if value == nil {
		return
	}
	return t.db.Model(&tables.Task{}).Where("name = ?", t.name).Updates(value).Error
}

func (t *task) GetState() (state int, err error) {
	err = t.db.Model(&tables.Task{}).Select("state").Where("name = ?", t.name).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return
}

func (t *task) Env() backend.IEnv {
	return &taskEnv{
		db:       t.db,
		taskName: t.name,
	}
}

type taskEnv struct {
	db       *gorm.DB
	taskName string
}

func (t *taskEnv) List() (res []*models.Env) {
	t.db.Model(&tables.TaskEnv{}).Where("task_name = ?", t.taskName).Order("id ASC").Find(&res)
	return
}

func (t *taskEnv) Create(env []*models.Env) (err error) {
	if len(env) == 0 {
		return
	}
	var envs []tables.TaskEnv
	for _, value := range env {
		envs = append(envs, tables.TaskEnv{
			TaskEnv: models.TaskEnv{
				TaskName: t.taskName,
				Env:      *value,
			},
		})
	}
	return t.db.Create(&envs).Error
}

func (t *taskEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = t.db.Model(&tables.TaskEnv{}).Select("value").Where("task_name = ? AND name = ?", t.taskName, name).First(res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.New("not found")
	}
	return
}

func (t *taskEnv) Delete(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return t.db.Where("task_name = ? AND name = ?", t.taskName, name).Delete(&tables.TaskEnv{}).Error
}

func (t *taskEnv) DeleteAll() (err error) {
	return t.db.Where("task_name = ?", t.taskName).Delete(&tables.TaskEnv{}).Error
}

func (t *task) Step(name string) backend.IStep {
	return &step{
		db:       t.db,
		taskName: t.name,
		name:     name,
	}
}

func (t *task) StepList(str string) (res []*models.Step) {
	if str != "" {
		t.db.Model(&tables.Step{}).Where("task_name = ? AND name LIKE ?", t.name, "%s"+str+"%s").Order("id ASC").Find(&res)
		return
	}
	t.db.Model(&tables.Step{}).Where("task_name = ?", t.name).Order("id ASC").Find(&res)
	return
}
