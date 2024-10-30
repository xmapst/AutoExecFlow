package touch

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

type STouch struct {
	storage   storage.IStep
	workspace string

	Path      string `json:"path" yaml:"Path"`
	Overwrite bool   `json:"overwrite" yaml:"Overwrite"`
	Content   string `json:"content" yaml:"Content"`
}

func New(
	storage storage.IStep,
	workspace string,
) (*STouch, error) {
	return &STouch{
		storage:   storage,
		workspace: workspace,
	}, nil
}

func (t *STouch) Run(ctx context.Context) (code int64, err error) {
	content, err := t.storage.Content()
	if err != nil {
		return common.CodeSystemErr, err
	}
	if err = json.Unmarshal([]byte(content), t); err != nil {
		if err = yaml.Unmarshal([]byte(content), t); err != nil {
			return common.CodeSystemErr, err
		}
	}
	t.Path = filepath.Clean(t.Path)
	if t.Path == "" {
		return common.CodeSystemErr, fmt.Errorf("path is empty")
	}
	if t.Overwrite {
		t.storage.Log().Writef("overwrite %s", t.Path)
		err = os.WriteFile(filepath.Join(t.workspace, t.Path), []byte(t.Content), os.ModePerm)
	} else {
		t.storage.Log().Writef("create or append %s", t.Path)
		file, err := os.OpenFile(filepath.Join(t.workspace, t.Path), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		defer file.Close()
		if err != nil {
			return common.CodeSystemErr, err
		}
		_, err = file.WriteString(t.Content)
	}
	if err != nil {
		return common.CodeSystemErr, err
	}
	return common.CodeSuccess, nil
}

func (t *STouch) Clear() error {
	return nil
}
