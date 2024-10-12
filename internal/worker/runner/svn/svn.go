package svn

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/storage"
)

type Svn struct {
}

func New(storage storage.IStep, command, workspace string) (*Svn, error) {
	return &Svn{}, nil
}

func (s *Svn) Run(ctx context.Context) (code int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *Svn) Clear() error {
	//TODO implement me
	panic("implement me")
}
