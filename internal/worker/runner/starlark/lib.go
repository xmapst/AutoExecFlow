package starlark

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/go-cmd/cmd"
	"go.starlark.net/starlark"
)

func (s *SStarLark) readFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var filename string
	if err := starlark.UnpackArgs("read_file", args, kwargs, "filename", &filename); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read_file: %v", err)
	}
	return starlark.String(content), nil
}

func (s *SStarLark) writeFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		filename string
		content  string
	)
	if err := starlark.UnpackArgs("write_file", args, kwargs, "filename", &filename, "content", &content); err != nil {
		return nil, err
	}
	return starlark.None, os.WriteFile(filename, []byte(content), os.ModePerm)
}
func (s *SStarLark) execCommand(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := thread.Local("star_ctx").(context.Context)
	var command string
	if err := starlark.UnpackArgs("exec_command", args, kwargs, "cmd", &command); err != nil {
		return nil, err
	}
	shell := cmd.NewCmd("sh", "-c", command)
	if runtime.GOOS == "windows" {
		shell = cmd.NewCmd("cmd", "/C", command)
	}
	defer func(shell *cmd.Cmd) {
		_ = shell.Stop()
	}(shell)
	shell.Dir = s.workspace
	shell.Env = os.Environ()

	select {
	case <-ctx.Done():
		_ = shell.Stop()
		return nil, fmt.Errorf("command execution canceled")
	case status := <-shell.Start():
		if status.Error != nil {
			return nil, fmt.Errorf("command execution failed: %w", status.Error)
		}

		return starlark.String(strings.Join(status.Stdout, "\n")), nil
	}
}
