package storage

import (
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sProjectBuild struct {
	*gorm.DB
	pName string
}

func (p *sProjectBuild) List() (res models.ProjectBuilds) {
	p.Model(&models.ProjectBuilds{}).
		Where(map[string]interface{}{
			"project_name": p.pName,
		}).
		Order("id ASC").
		Find(&res)
	return
}

func (p *sProjectBuild) Task(name string) ITask {
	return &sTask{
		DB:    p.DB,
		tName: name,
	}
}

func (p *sProjectBuild) Get(name string) (res *models.ProjectBuild, err error) {
	res = new(models.ProjectBuild)
	err = p.Where(map[string]interface{}{
		"project_name": p.pName,
		"task_name":    name,
	}).First(res).Error
	return
}

func (p *sProjectBuild) Insert(names ...string) (err error) {
	if len(names) == 0 {
		return
	}
	var tasks []models.ProjectBuild
	for _, name := range names {
		tasks = append(tasks, models.ProjectBuild{
			ProjectName: p.pName,
			TaskName:    name,
		})
	}
	return p.Create(&tasks).Error
}

func (p *sProjectBuild) Remove(name string) (err error) {
	return p.Where(map[string]interface{}{
		"project_name": p.pName,
		"task_name":    name,
	}).Delete(&models.ProjectBuild{}).Error
}

func (p *sProjectBuild) ClearAll() error {
	return p.Where("project_name = ?", p.pName).Delete(&models.ProjectBuild{}).Error
}
