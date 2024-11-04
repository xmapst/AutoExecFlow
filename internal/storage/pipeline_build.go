package storage

import (
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sPipelineBuild struct {
	*gorm.DB
	pName string
}

func (p *sPipelineBuild) List(page, size int64) (res models.SPipelineBuilds, total int64) {
	query := p.Table("t_pipeline_build p").
		Select("p.id AS id, p.pipeline_name AS pipeline_name, p.task_name AS task_name, " +
			"t.state AS state, t.message AS message, t.s_time AS  s_time, t.e_time AS e_time").
		Joins("INNER JOIN t_task t ON t.name = p.task_name").
		Where(map[string]interface{}{
			"pipeline_name": p.pName,
		}).
		Order("id DESC").
		Count(&total)
	if page <= 0 || size <= 0 {
		query.Find(&res)
		return
	}
	query.Scopes(func(db *gorm.DB) *gorm.DB {
		return models.Paginate(db, page, size)
	}).Find(&res)
	return
}

func (p *sPipelineBuild) Get(name string) (res *models.SPipelineBuildRes, err error) {
	err = p.Table("t_pipeline_build p").
		Select("p.id AS id, p.pipeline_name AS pipeline_name, p.task_name AS task_name, p.params AS params, " +
			"t.state AS state, t.message AS message, t.s_time AS  s_time, t.e_time AS e_time").
		Joins("INNER JOIN t_task t ON t.name = p.task_name").
		Where(map[string]interface{}{
			"pipeline_name": p.pName,
			"task_name":     name,
		}).Find(&res).Error
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
