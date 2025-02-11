package mkdir

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

type SMkdir struct {
	storage   storage.IStep
	workspace string

	Path string `json:"path" yaml:"Path"` // 文件夹路径
}

func New(
	storage storage.IStep,
	workspace string,
) (*SMkdir, error) {
	return &SMkdir{
		storage:   storage,
		workspace: workspace,
	}, nil
}

func (m *SMkdir) Run(ctx context.Context) (exit int64, err error) {
	content, err := m.storage.Content()
	if err != nil {
		return common.CodeSystemErr, err
	}
	if err = json.Unmarshal([]byte(content), m); err != nil {
		if err = yaml.Unmarshal([]byte(content), m); err != nil {
			return common.CodeSystemErr, err
		}
	}
	m.Path = filepath.Clean(m.Path)
	if m.Path == "" {
		return common.CodeSystemErr, fmt.Errorf("path is empty")
	}
	m.storage.Log().Writef("mkdir -p %s", m.Path)
	err = os.MkdirAll(filepath.Join(m.workspace, m.Path), os.ModePerm)
	if err != nil {
		return common.CodeSystemErr, err
	}
	return common.CodeSuccess, nil
}

func (m *SMkdir) Clear() error {
	return nil
}
