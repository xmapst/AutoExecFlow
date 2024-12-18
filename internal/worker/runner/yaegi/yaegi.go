package yaegi

import (
	"context"
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"github.com/traefik/yaegi/stdlib/syscall"
	"github.com/traefik/yaegi/stdlib/unrestricted"
	"github.com/traefik/yaegi/stdlib/unsafe"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

type SYaegi struct {
	interp    *interp.Interpreter
	storage   storage.IStep
	workspace string
}

func New(storage storage.IStep, workspace string) (*SYaegi, error) {
	return &SYaegi{
		storage:   storage,
		workspace: workspace,
	}, nil
}

func (y *SYaegi) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		if _r := recover(); _r != nil {
			err = fmt.Errorf("panic during execution %v", _r)
			code = common.CodeSystemErr
			stack := debug.Stack()
			if _err, ok := _r.(error); ok && strings.Contains(_err.Error(), context.Canceled.Error()) {
				code = common.CodeKilled
				err = common.ErrManual
			}
			y.storage.Log().Write(err.Error(), string(stack))
		}
	}()

	content, err := y.storage.Content()
	if err != nil {
		return common.CodeFailed, err
	}

	var params = map[string]any{}
	taskEnv := y.storage.GlobalEnv().List()
	for _, v := range taskEnv {
		params[v.Name] = v.Value
	}
	stepEnv := y.storage.Env().List()
	for _, v := range stepEnv {
		params[v.Name] = v.Value
	}

	if err = y.createVM(); err != nil {
		return common.CodeSystemErr, err
	}

	_, err = y.interp.EvalWithContext(ctx, content)
	if err != nil {
		return common.CodeFailed, err
	}

	evalFnval, err := y.interp.EvalWithContext(ctx, "EvalCall")
	if err != nil {
		return common.CodeFailed, err
	}
	evalFn, ok := evalFnval.Interface().(func(map[string]any))
	if !ok {
		return common.CodeFailed, errors.New("not found EvalCall")
	}
	evalFn(params)
	return common.CodeSuccess, nil
}

func (y *SYaegi) createVM() (err error) {
	y.interp = interp.New(interp.Options{
		Env: []string{
			fmt.Sprintf("WORKSPACE=%s", y.workspace),
		},
		Stdout: y.output(),
		Stderr: y.output(),
	})

	_ = y.interp.Use(stdlib.Symbols)
	_ = y.interp.Use(unsafe.Symbols)
	_ = y.interp.Use(syscall.Symbols)
	_ = y.interp.Use(unrestricted.Symbols)
	_ = y.interp.Use(interp.Symbols)

	return
}

func (y *SYaegi) Clear() error {
	return nil
}

type sYeagiOutput struct {
	storage storage.IStep
}

func (s *sYeagiOutput) Write(p []byte) (n int, err error) {
	n = len(p)
	s.storage.Log().Write(string(p))
	return
}

func (y *SYaegi) output() io.Writer {
	return &sYeagiOutput{
		storage: y.storage,
	}
}
