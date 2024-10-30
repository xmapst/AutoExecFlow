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

	"github.com/xmapst/AutoExecFlow/internal/worker/plugins"
)

func init() {
	plugins.Register(Name, new(SPlugin))
}

const Name = "yeagi"

type SPlugin struct {
	env []string
}

func (p *SPlugin) Name() string {
	return Name
}

func (p *SPlugin) Description() string {
	return "yeagi plugin provider"
}

func (p *SPlugin) WithEnv(env []string) plugins.IPlugin {
	p.env = append(p.env, env...)
	return p
}

func (p *SPlugin) Run(ctx context.Context, content string) error {
	filename := filepath.Join(os.TempDir(), ksuid.New().String())
	defer os.RemoveAll(filename)
	if err := os.WriteFile(filename, []byte(content), os.ModePerm); err != nil {
		return err
	}
	return p.RunFile(ctx, filename)
}

func (p *SPlugin) RunFile(ctx context.Context, filename string) error {
	vm := interp.New(interp.Options{
		Env: p.env,
	})
	if err := vm.Use(stdlib.Symbols); err != nil {
		return fmt.Errorf("failed to load Go runtime: %v", err)
	}
	if err := vm.Use(stdlib.Symbols); err != nil {
		return fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err := vm.Use(unsafe.Symbols); err != nil {
		return fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err := vm.Use(syscall.Symbols); err != nil {
		return fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err := vm.Use(unrestricted.Symbols); err != nil {
		return fmt.Errorf("failed to load Go runtime: %v", err)
	}

	if err := vm.Use(interp.Symbols); err != nil {
		return fmt.Errorf("failed to load Go runtime: %v", err)
	}

	return p.executeSafely(ctx, vm, filename)
}

func (p *SPlugin) executeSafely(ctx context.Context, vm *interp.Interpreter, filename string) (err error) {
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
		return err
	}
	_, err = vm.ExecuteWithContext(ctx, prog)
	if err != nil {
		return err
	}

	return nil
}
