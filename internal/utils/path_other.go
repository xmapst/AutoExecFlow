//go:build !windows

package utils

import (
	"fmt"
	"path/filepath"
)

var DefaultDir = filepath.Join("/", "usr", "local", ServiceName)

var PipeName = fmt.Sprintf("unix:///var/run/%s.sock", ServiceName)
