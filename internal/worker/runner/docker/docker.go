package docker

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/storage"
)

type SDocker struct {
}

func New(storage storage.IStep, command, workspace string) (*SDocker, error) {
	return &SDocker{}, nil
}

func (d *SDocker) Run(ctx context.Context) (code int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (d *SDocker) Clear() error {
	//TODO implement me
	panic("implement me")
}
