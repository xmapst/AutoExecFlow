package service

import (
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/types"
)

type SProjectService struct {
	name string
}

func Project(name string) *SProjectService {
	return &SProjectService{
		name: name,
	}
}

func ProjectList(req *types.SPageReq) *types.SProjectListRes {
	projects, total := storage.ProjectList(req.Page, req.Size, req.Prefix)
	if projects == nil {
		return nil
	}
	pageTotal := total / req.Size
	if total%req.Size != 0 {
		pageTotal += 1
	}
	var list = &types.SProjectListRes{
		Page: types.SPageRes{
			Current: req.Page,
			Size:    req.Size,
			Total:   pageTotal,
		},
	}
	for _, project := range projects {
		list.Projects = append(list.Projects, &types.SProjectRes{
			Name:    project.Name,
			Disable: *project.Disable,
			Type:    project.Type,
		})
	}
	return list
}

func (p *SProjectService) Delete() error {
	return storage.Project(p.name).ClearAll()
}

func (p *SProjectService) Detail() (*types.SProjectRes, error) {
	res, err := storage.Project(p.name).Get()
	if err != nil {
		return nil, err
	}
	return &types.SProjectRes{
		Name:        res.Name,
		Description: res.Description,
		Disable:     *res.Disable,
		Type:        res.Type,
		Content:     res.Content,
	}, nil
}

func (p *SProjectService) Create(req *types.SProjectCreateReq) error {
	return storage.ProjectCreate(&models.SProject{
		Name: p.name,
		SProjectUpdate: models.SProjectUpdate{
			Description: req.Description,
			Disable:     req.Disable,
			Type:        req.Type,
			Content:     req.Content,
		},
	})
}

func (p *SProjectService) Update(req *types.SProjectUpdateReq) error {
	return storage.Project(p.name).Update(&models.SProjectUpdate{
		Description: req.Description,
		Disable:     req.Disable,
		Type:        req.Type,
		Content:     req.Content,
	})
}

func (p *SProjectService) BuildList() interface{} {
	return nil
}
