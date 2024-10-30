package service

import (
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/types"
)

type SPoolService struct {
}

func Pool() *SPoolService {
	return &SPoolService{}
}

func (p *SPoolService) Get() *types.SPool {
	return &types.SPool{
		Size:    worker.GetSize(),
		Total:   storage.TaskCount(models.StateAll),
		Running: storage.TaskCount(models.StateRunning),
		Waiting: storage.TaskCount(models.StatePending),
	}
}

func (p *SPoolService) Set(size int) (*types.SPool, error) {
	if size <= 0 {
		return p.Get(), nil
	}
	if (storage.TaskCount(models.StateRunning) != 0 || storage.TaskCount(models.StatePending) != 0) && size <= worker.GetSize() {
		return nil, errors.New("there are still tasks running, scaling down is not allowed")
	}
	worker.SetSize(size)
	return p.Get(), nil
}
