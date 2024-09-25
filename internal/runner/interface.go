package runner

import (
	"context"
)

type IRunner interface {
	Run(ctx context.Context) (code int64, err error)
	Clear() error
}
