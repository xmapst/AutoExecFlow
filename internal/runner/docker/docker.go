package docker

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/storage"
)

type Docker struct {
}

func New(storage storage.IStep, command, workspace string) (*Docker, error) {
	return &Docker{}, nil
}

func (d *Docker) Run(ctx context.Context) (code int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (d *Docker) Clear() error {
	//TODO implement me
	panic("implement me")
}
