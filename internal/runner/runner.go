package runner

import (
	"strings"

	"github.com/xmapst/AutoExecFlow/internal/runner/docker"
	"github.com/xmapst/AutoExecFlow/internal/runner/exec"
	"github.com/xmapst/AutoExecFlow/internal/runner/git"
	"github.com/xmapst/AutoExecFlow/internal/runner/kubernetes"
	"github.com/xmapst/AutoExecFlow/internal/runner/mkdir"
	"github.com/xmapst/AutoExecFlow/internal/runner/touch"
	"github.com/xmapst/AutoExecFlow/internal/storage"
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
		return kubernetes.New(storage, strings.TrimPrefix(commandType, "kubectl@"), workspace)
	case strings.EqualFold(commandType, "mkdir"):
		return mkdir.New(storage, workspace)
	case strings.EqualFold(commandType, "touch"):
		return touch.New(storage, workspace)
	default:
		return exec.New(storage, commandType, workspace, scriptDir)
	}
}
