package sqlite

import (
	"gorm.io/gorm"

	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
)

type stepDepend struct {
	db    *gorm.DB
	tName string
	sName string
}

func (s *stepDepend) List() (res []string) {
	s.db.Model(&tables.StepDepend{}).Select("name").Where("task_name = ? AND step_name = ?", s.tName, s.sName).Order("id ASC").Find(&res)
	return
}

func (s *stepDepend) Create(depends ...string) (err error) {
	if len(depends) == 0 {
		return
	}
	var stepDepends []tables.StepDepend
	for _, depend := range depends {
		stepDepends = append(stepDepends, tables.StepDepend{
			TaskName: s.tName,
			StepName: s.sName,
			Name:     depend,
		})
	}
	return s.db.Create(&stepDepends).Error
}

func (s *stepDepend) DeleteAll() (err error) {
	return s.db.Where("task_name = ? AND step_name = ?", s.tName, s.sName).Delete(&tables.StepDepend{}).Error
}
