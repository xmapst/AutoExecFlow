package runner

import (
	"strings"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/docker"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/exec"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/git"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/k8s"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/lua"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/mkdir"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/starlark"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/touch"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner/yaegi"
)

func New(
	storage storage.IStep,
	workspace, scriptDir string,
) (IRunner, error) {
	commandType, err := storage.Type()
	if err != nil {
		return nil, err
	}
	switch {
	case strings.HasPrefix(commandType, "docker"):
		return docker.New(storage, strings.TrimPrefix(commandType, "docker@"), workspace)
	case strings.HasPrefix(commandType, "git"):
		return git.New(storage, workspace)
	case strings.HasPrefix(commandType, "kubectl"):
		return k8s.New(storage, strings.TrimPrefix(commandType, "kubectl@"), workspace)
	case strings.EqualFold(commandType, "mkdir"):
		return mkdir.New(storage, workspace)
	case strings.EqualFold(commandType, "touch"):
		return touch.New(storage, workspace)
	case strings.EqualFold(commandType, "lua"):
		return lua.New(storage, workspace)
	case strings.EqualFold(commandType, "star"):
		return starlark.New(storage, workspace)
	case strings.EqualFold(commandType, "yaegi"):
		return yaegi.New(storage, workspace)
	default:
		return exec.New(storage, commandType, workspace, scriptDir)
	}
}
