package yeagi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/segmentio/ksuid"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/syscall"
	"github.com/traefik/yaegi/stdlib/unrestricted"
	"github.com/traefik/yaegi/stdlib/unsafe"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

type SYeagi struct {
	vm        *interp.Interpreter
	storage   storage.IStep
	workspace string
	scriptDir string
}

func New(storage storage.IStep, workspace, scriptDir string) (*SYeagi, error) {
	return &SYeagi{
		storage:   storage,
		workspace: workspace,
		scriptDir: scriptDir,
	}, nil
}

func (y *SYeagi) Run(ctx context.Context) (code int64, err error) {
	filename := filepath.Join(os.TempDir(), ksuid.New().String())
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(filename)
	content, err := y.storage.Content()
	if err != nil {
		return common.CodeFailed, err
	}
	if err = os.WriteFile(filename, []byte(content), os.ModePerm); err != nil {
		return common.CodeFailed, err
	}

	var params = map[string]string{}
	taskEnv := y.storage.GlobalEnv().List()
	for _, v := range taskEnv {
		params[v.Name] = v.Value
	}
	stepEnv := y.storage.Env().List()
	for _, v := range stepEnv {
		params[v.Name] = v.Value
	}

	vm := interp.New(interp.Options{
		Env: utils.MapToSlice(params),
	})

	if err = vm.Use(stdlib.Symbols); err != nil {
		return common.CodeFailed, fmt.Errorf("failed to load Go runtime: %v", err)
	}
	if err = vm.Use(stdlib.Symbols); err != nil {
		return common.CodeFailed, fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err = vm.Use(unsafe.Symbols); err != nil {
		return common.CodeFailed, fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err = vm.Use(syscall.Symbols); err != nil {
		return common.CodeFailed, fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err = vm.Use(unrestricted.Symbols); err != nil {
		return common.CodeFailed, fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err = vm.Use(interp.Symbols); err != nil {
		return common.CodeFailed, fmt.Errorf("failed to load Go runtime: %v", err)
	}
	defer func() {
		if _r := recover(); _r != nil {
			if err != nil {
				err = fmt.Errorf("panic during execution %v %v", err, _r)
				return
			}
			err = fmt.Errorf("panic during execution %v", _r)
		}
	}()
	prog, err := vm.CompilePath(filename)
	if err != nil {
		return common.CodeFailed, err
	}
	_, err = vm.ExecuteWithContext(ctx, prog)
	if err != nil {
		return common.CodeFailed, err
	}

	return common.CodeSuccess, nil
}

func (y *SYeagi) Clear() error {
	return nil
}
