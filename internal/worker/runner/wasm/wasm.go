package wasm

import (
	"context"

	"github.com/xmapst/AutoExecFlow/internal/storage"
)

type SWasm struct {
	storage   storage.IStep
	workspace string
}

func New(storage storage.IStep, workspace string) (*SWasm, error) {
	return &SWasm{
		storage:   storage,
		workspace: workspace,
	}, nil
}

func (w *SWasm) Run(ctx context.Context) (exit int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (w *SWasm) Clear() error {
	//TODO implement me
	panic("implement me")
}
