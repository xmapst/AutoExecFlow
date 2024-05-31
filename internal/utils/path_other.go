//go:build !windows

package utils

import (
	"path/filepath"
)

var DefaultDir = filepath.Join("/", "usr", "local", ServiceName)
