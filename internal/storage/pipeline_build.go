package storage

import (
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sPipelineBuild struct {
	*gorm.DB
	pName string
}

func (p *sPipelineBuild) List(page, size int64) (res []string) {
	query := p.Model(&models.SPipelineBuilds{}).
		Select("task_name").
		Where(map[string]interface{}{
			"pipeline_name": p.pName,
		}).
		Order("id DESC")
	if page <= 0 || size <= 0 {
		query.Find(&res)
		return
	}
	query.Scopes(func(db *gorm.DB) *gorm.DB {
		return models.Paginate(db, page, size)
	}).Find(&res)
	return
}

func (p *sPipelineBuild) Task(name string) ITask {
	return &sTask{
		DB:    p.DB,
		tName: name,
	}
}

func (p *sPipelineBuild) Get(name string) (res *models.SPipelineBuild, err error) {
	res = new(models.SPipelineBuild)
	err = p.Where(map[string]interface{}{
		"pipeline_name": p.pName,
		"task_name":     name,
	}).First(res).Error
	return
}

func (p *sPipelineBuild) Insert(build *models.SPipelineBuild) (err error) {
	build.PipelineName = p.pName
	return p.Create(&build).Error
}

func (p *sPipelineBuild) Remove(name string) (err error) {
	return p.Where(map[string]interface{}{
		"pipeline_name": p.pName,
		"task_name":     name,
	}).Delete(&models.SPipelineBuild{}).Error
}

func (p *sPipelineBuild) ClearAll() error {
	return p.Where("pipeline_name = ?", p.pName).Delete(&models.SPipelineBuild{}).Error
}
