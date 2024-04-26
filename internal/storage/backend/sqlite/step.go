package sqlite

import (
	"errors"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type step struct {
	db       *gorm.DB
	taskName string
	name     string
}

func (s *step) ClearAll() {
	_ = s.Delete()
	_ = s.Env().DeleteAll()
	_ = s.Depend().DeleteAll()
}

func (s *step) Get() (res *models.Step, err error) {
	res = new(models.Step)
	err = s.db.Model(&tables.Step{}).Where("task_name = ? AND name = ?", s.taskName, s.name).First(res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("not found")
	}
	return
}

func (s *step) Delete() (err error) {
	err = s.db.Where("task_name = ? AND name = ?", s.taskName, s.name).Delete(&tables.Step{}).Error
	return
}

func (s *step) Create(step *models.Step) (err error) {
	step.TaskName = s.taskName
	step.Name = s.name
	err = s.db.Create(&tables.Step{
		Step: *step,
	}).Error
	if err != nil {
		return err
	}
	return
}

func (s *step) Update(value *models.StepUpdate) (err error) {
	if value == nil {
		return
	}
	return s.db.Model(&tables.Step{}).Where("task_name = ? AND name = ?", s.taskName, s.name).Updates(value).Error
}

func (s *step) GetState() (state int, err error) {
	err = s.db.Model(&tables.Step{}).Select("state").Where("task_name = ? AND name = ?", s.taskName, s.name).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return
}

func (s *step) Depend() backend.IDepend {
	return &stepDepend{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

type stepDepend struct {
	db       *gorm.DB
	taskName string
	stepName string
}

func (d *stepDepend) List() (res []string) {
	d.db.Model(&tables.StepDepend{}).Select("name").Where("task_name = ? AND step_name = ?", d.taskName, d.stepName).Order("id ASC").Find(&res)
	return
}

func (d *stepDepend) Create(depends []string) (err error) {
	if len(depends) == 0 {
		return
	}
	var stepDepends []tables.StepDepend
	for _, depend := range depends {
		stepDepends = append(stepDepends, tables.StepDepend{
			StepDepend: models.StepDepend{
				TaskName: d.taskName,
				StepName: d.stepName,
				Name:     depend,
			},
		})
	}
	return d.db.Create(&stepDepends).Error
}

func (d *stepDepend) DeleteAll() (err error) {
	return d.db.Where("task_name = ? AND step_name = ?", d.taskName, d.stepName).Delete(&tables.StepDepend{}).Error
}

func (s *step) Env() backend.IEnv {
	return &stepEnv{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

type stepEnv struct {
	db       *gorm.DB
	taskName string
	stepName string
}

func (s *stepEnv) List() (res []*models.Env) {
	s.db.Model(&tables.StepEnv{}).Where("task_name = ? AND step_name = ?", s.taskName, s.stepName).Order("id ASC").Find(&res)
	return
}

func (s *stepEnv) Create(env []*models.Env) (err error) {
	if len(env) == 0 {
		return
	}
	var envs []tables.StepEnv
	for _, value := range env {
		envs = append(envs, tables.StepEnv{
			StepEnv: models.StepEnv{
				TaskName: s.taskName,
				StepName: s.stepName,
				Env:      *value,
			},
		})
	}
	return s.db.Create(&envs).Error
}

func (s *stepEnv) Get(name string) (res string, err error) {
	if name == "" {
		return "", errors.New("name is empty")
	}
	err = s.db.Model(&tables.StepEnv{}).Select("value").Where("task_name = ? AND step_name = ? AND name = ?", s.taskName, s.stepName, name).First(res).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.New("not found")
	}
	return
}

func (s *stepEnv) Delete(name string) (err error) {
	if name == "" {
		return errors.New("name is empty")
	}
	return s.db.Where("task_name = ? AND step_name = ? AND name = ?", s.taskName, s.stepName, name).Delete(&tables.StepEnv{}).Error
}

func (s *stepEnv) DeleteAll() (err error) {
	return s.db.Where("task_name = ? AND step_name = ?", s.taskName, s.stepName).Delete(&tables.StepEnv{}).Error
}

func (s *step) Log() backend.ILog {
	return &log{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}
