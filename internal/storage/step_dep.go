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

func (s *sStepDepend) List() (res []string) {
	s.Model(&models.SStepDepend{}).
		Select("name").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"step_name": s.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (s *sStepDepend) Insert(depends ...string) (err error) {
	if len(depends) == 0 {
		return
	}
	var stepDepends []models.SStepDepend
	for _, depend := range depends {
		stepDepends = append(stepDepends, models.SStepDepend{
			TaskName: s.tName,
			StepName: s.sName,
			Name:     depend,
		})
	}
	return s.Create(&stepDepends).Error
}

func (s *sStepDepend) RemoveAll() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"step_name": s.sName,
	}).Delete(&models.SStepDepend{}).Error
}
