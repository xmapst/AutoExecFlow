package svn

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/storage"
)

type SSvn struct {
}

func New(storage storage.IStep, command, workspace string) (*SSvn, error) {
	return &SSvn{}, nil
}

func (s *SSvn) Run(ctx context.Context) (code int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *SSvn) Clear() error {
	//TODO implement me
	panic("implement me")
}
