package storage

import (
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
)

type sProject struct {
	*gorm.DB
	name string
}

func (p *sProject) Name() string {
	return p.name
}

func (p *sProject) ClearAll() error {
	if err := p.Remove(); err != nil {
		return err
	}
	builds := p.Build().List()
	for _, build := range builds {
		if err := p.Task(build.TaskName).ClearAll(); err != nil {
			return err
		}
	}
	return p.Build().ClearAll()
}

func (p *sProject) Remove() (err error) {
	return p.Where(map[string]interface{}{
		"name": p.name,
	}).Delete(&models.SProject{}).Error
}

func (p *sProject) Build() IProjectBuild {
	return &sProjectBuild{
		pName: p.name,
		DB:    p.DB,
	}
}

func (p *sProject) Task(name string) ITask {
	return &sTask{
		DB:    p.DB,
		tName: name,
	}
}

func (p *sProject) Update(value *models.SProjectUpdate) (err error) {
	if value == nil {
		return
	}
	return p.Model(&models.SProject{}).
		Where(map[string]interface{}{
			"name": p.name,
		}).
		Updates(value).
		Error
}

func (p *sProject) Get() (res *models.SProject, err error) {
	res = new(models.SProject)
	err = p.Where(map[string]interface{}{
		"name": p.name,
	}).First(res).Error
	return
}

func (p *sProject) IsDisable() (disable bool) {
	if p.Model(&models.SProject{}).
		Select("disable").
		Where(map[string]interface{}{
			"name": p.name,
		}).
		Scan(&disable).
		Error != nil {
		return
	}
	return
}

func (p *sProject) Type() (res string, err error) {
	err = p.Model(&models.SProject{}).
		Select("type").
		Where(map[string]interface{}{
			"name": p.name,
		}).
		Scan(&res).
		Error
	return
}

func (p *sProject) Content() (res string, err error) {
	err = p.Model(&models.SProject{}).
		Select("content").
		Where(map[string]interface{}{
			"name": p.name,
		}).
		Scan(&res).
		Error
	return
}
