package storage

import (
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sStepDepend struct {
	*gorm.DB
	tName string
	sName string
}

func (d *sStepDepend) List() (res []string) {
	d.Model(&models.SStepDepend{}).
		Select("name").
		Where(map[string]interface{}{
			"task_name": d.tName,
			"step_name": d.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (d *sStepDepend) Insert(depends ...string) (err error) {
	if len(depends) == 0 {
		return
	}
	var stepDepends []models.SStepDepend
	for _, depend := range depends {
		stepDepends = append(stepDepends, models.SStepDepend{
			TaskName: d.tName,
			StepName: d.sName,
			Name:     depend,
		})
	}
	return d.Create(&stepDepends).Error
}

func (d *sStepDepend) RemoveAll() (err error) {
	return d.Where(map[string]interface{}{
		"task_name": d.tName,
		"step_name": d.sName,
	}).Delete(&models.SStepDepend{}).Error
}
