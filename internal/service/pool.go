package service

import (
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/types"
)

type PoolService struct {
}

func Pool() *PoolService {
	return &PoolService{}
}

func (p *PoolService) Get() *types.Pool {
	return &types.Pool{
		Size:    worker.GetSize(),
		Total:   worker.GetTotal(),
		Running: worker.Running(),
		Waiting: worker.Waiting(),
	}
}

func (p *PoolService) Set(size int) (*types.Pool, error) {
	if size <= 0 {
		return p.Get(), nil
	}
	if (worker.Running() != 0 || worker.Waiting() != 0) && size <= worker.GetSize() {
		return nil, errors.New("there are still tasks running, scaling down is not allowed")
	}
	worker.SetSize(size)
	return p.Get(), nil
}
