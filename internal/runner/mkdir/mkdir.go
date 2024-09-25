package mkdir

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
)

type Mkdir struct {
	storage   backend.IStep
	workspace string

	Path string `json:"path" yaml:"Path"` // 文件夹路径
}

func New(
	storage backend.IStep,
	workspace string,
) (*Mkdir, error) {
	return &Mkdir{
		storage:   storage,
		workspace: workspace,
	}, nil
}

func (m *Mkdir) Run(ctx context.Context) (code int64, err error) {
	content, err := m.storage.Content()
	if err != nil {
		return common.SystemErr, err
	}
	if err = json.Unmarshal([]byte(content), m); err != nil {
		if err = yaml.Unmarshal([]byte(content), m); err != nil {
			return common.SystemErr, err
		}
	}
	m.Path = filepath.Clean(m.Path)
	if m.Path == "" {
		return common.SystemErr, fmt.Errorf("path is empty")
	}
	m.storage.Log().Writef("mkdir -p %s", m.Path)
	err = os.MkdirAll(filepath.Join(m.workspace, m.Path), os.ModePerm)
	if err != nil {
		return common.SystemErr, err
	}
	return common.Success, nil
}

func (m *Mkdir) Clear() error {
	return nil
}
