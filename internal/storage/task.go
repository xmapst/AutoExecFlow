package storage

import (
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sTask struct {
	*gorm.DB
	tName string

	env IEnv
}

func (t *sTask) Name() string {
	return t.tName
}

func (t *sTask) ClearAll() error {
	if err := t.Remove(); err != nil {
		return err
	}
	if err := t.Env().RemoveAll(); err != nil {
		return err
	}
	list := t.StepList(All)
	for _, v := range list {
		if err := t.Step(v.Name).ClearAll(); err != nil {
			return err
		}
	}
	// 清理build表
	t.Where("task_name", t.tName).Delete(&models.SPipelineBuild{})
	return nil
}

func (t *sTask) Remove() (err error) {
	return t.Where(map[string]interface{}{
		"name": t.tName,
	}).Delete(&models.STask{}).Error
}

func (t *sTask) State() (state models.State, err error) {
	err = t.Model(&models.STask{}).
		Select("state").
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Scan(&state).
		Error
	return
}

func (t *sTask) IsDisable() (disable bool) {
	if t.Model(&models.STask{}).
		Select("disable").
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Scan(&disable).
		Error != nil {
		return
	}
	return
}

func (t *sTask) Env() IEnv {
	if t.env == nil {
		t.env = &sTaskEnv{
			DB:    t.DB,
			tName: t.tName,
		}
	}
	return t.env
}

func (t *sTask) Timeout() (res time.Duration, err error) {
	err = t.Model(&models.STask{}).
		Select("timeout").
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Scan(&res).
		Error
	return
}

func (t *sTask) Get() (res *models.STask, err error) {
	res = new(models.STask)
	err = t.Model(&models.STask{}).
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		First(res).
		Error
	return
}

func (t *sTask) Update(value *models.STaskUpdate) (err error) {
	if value == nil {
		return
	}
	return t.Model(&models.STask{}).
		Where(map[string]interface{}{
			"name": t.tName,
		}).
		Updates(value).
		Error
}

func (t *sTask) Step(name string) IStep {
	return &sStep{
		DB:    t.DB,
		genv:  t.Env(),
		tName: t.tName,
		sName: name,
	}
}

func (t *sTask) StepCreate(step *models.SStep) (err error) {
	step.TaskName = t.tName
	return t.Create(step).Error
}

func (t *sTask) StepCount() (res int64) {
	t.Model(&models.SStep{}).Count(&res)
	return
}

func (t *sTask) StepNameList(str string) (res []string) {
	query := t.Model(&models.SStep{}).
		Select("name").
		Order("id ASC").
		Where(map[string]interface{}{
			"task_name": t.tName,
		})
	if str != "" {
		query.Where("name LIKE ?", str)
	}
	query.Find(&res)
	return
}

func (t *sTask) StepStateList(str string) (res map[string]models.State) {
	var steps models.SSteps
	query := t.Model(&models.SStep{}).
		Select("name, state").
		Order("id ASC").
		Where(map[string]interface{}{
			"task_name": t.tName,
		})
	if str != "" {
		query.Where("name LIKE ?", str)
	}
	query.Find(&steps)
	res = make(map[string]models.State, len(steps))
	for _, v := range steps {
		res[v.Name] = *v.State
	}
	return
}

func (t *sTask) StepList(str string) (res models.SSteps) {
	query := t.Model(&models.SStep{}).
		Order("id ASC").
		Where(map[string]interface{}{
			"task_name": t.tName,
		})
	if str != "" {
		query.Where("name LIKE ?", str)
	}
	query.Find(&res)
	return
}

// CheckDependentModel 获取依赖当前步骤的步骤
func (t *sTask) CheckDependentModel() (res bool) {
	var count int64
	t.Table("t_step").Select("count(DISTINCT t_step.name)").
		Joins("INNER JOIN t_step_depend ON t_step.task_name = t_step_depend.task_name").
		Where("t_step.task_name = ? AND t_step.action != '' AND t_step.rule != ''", t.tName).
		Count(&count)
	return count > 0
}
