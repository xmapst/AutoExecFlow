package svn

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
)

type svn struct {
}

func New(storage backend.IStep, command, workspace string) (common.IRunner, error) {
	return &svn{}, nil
}

func (s *svn) Run(ctx context.Context) (code int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *svn) Clear() error {
	//TODO implement me
	panic("implement me")
}
