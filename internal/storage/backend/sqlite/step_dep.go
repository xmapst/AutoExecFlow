package sqlite

import (
	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type stepDepend struct {
	*gorm.DB
	tName string
	sName string
}

func (s *stepDepend) List() (res []string) {
	s.Model(&tables.StepDepend{}).
		Select("name").
		Where(map[string]interface{}{
			"task_name": s.tName,
			"step_name": s.sName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (s *stepDepend) Insert(depends ...string) (err error) {
	if len(depends) == 0 {
		return
	}
	var stepDepends []tables.StepDepend
	for _, depend := range depends {
		stepDepends = append(stepDepends, tables.StepDepend{
			TaskName: s.tName,
			StepName: s.sName,
			StepDepend: models.StepDepend{
				Name: depend,
			},
		})
	}
	return s.Create(&stepDepends).Error
}

func (s *stepDepend) RemoveAll() (err error) {
	return s.Where(map[string]interface{}{
		"task_name": s.tName,
		"step_name": s.sName,
	}).Delete(&tables.StepDepend{}).Error
}
