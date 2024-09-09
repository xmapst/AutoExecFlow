package docker

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
)

type docker struct {
}

func New(storage backend.IStep, command, workspace string) (common.IRunner, error) {
	return &docker{}, nil
}

func (d *docker) Run(ctx context.Context) (code int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (d *docker) Clear() error {
	//TODO implement me
	panic("implement me")
}
