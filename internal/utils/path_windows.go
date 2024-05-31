//go:build windows

package utils

import (
	"path/filepath"
)

var DefaultDir = filepath.Join("C:\\", "ProgramData", ServiceName)
