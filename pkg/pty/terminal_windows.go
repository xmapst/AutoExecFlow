//go:build windows

package pty

import (
	"github.com/xmapst/osreapi/pkg/pty/winpty"
)

type terminal struct {
	*winpty.ConPty
}

func New(cmd string) (Terminal, error) {
	if cmd == "" {
		cmd = "cmd.exe"
	}
	pty, err := winpty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &terminal{
		ConPty: pty,
	}, nil
}
