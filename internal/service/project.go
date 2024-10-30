package service

import "github.com/xmapst/AutoExecFlow/types"

type SProjectService struct {
	name string
}

func Project(name string) *SProjectService {
	return &SProjectService{
		name: name,
	}
}

func ProjectList(req *types.SPageReq) *types.SProjectListRes {
	return nil
}
