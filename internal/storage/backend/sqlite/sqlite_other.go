//go:build !amd64 && !arm64

package sqlite

import (
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/bbolt"
)

func New(path string) (backend.IStorage, error) {
	return bbolt.New(path)
}
