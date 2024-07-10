//go:build windows

package utils

import (
	"fmt"
	"path/filepath"
)

var DefaultDir = filepath.Join("C:\\", "ProgramData", ServiceName)

var PipeName = fmt.Sprintf("npipe:////./pipe/%s", ServiceName)
