package starlark

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
	"github.com/qri-io/starlib"
	"github.com/segmentio/ksuid"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	star_libs "github.com/xmapst/AutoExecFlow/pkg/star-libs"
)

type SStarLark struct {
	thread    *starlark.Thread
	storage   storage.IStep
	name      string
	workspace string
}

func New(storage storage.IStep, workspace string) (*SStarLark, error) {
	s := &SStarLark{
		name:      ksuid.New().String(),
		storage:   storage,
		workspace: workspace,
	}

	s.thread = &starlark.Thread{
		Name: s.name,
		Load: starlib.Loader,
		Print: func(thread *starlark.Thread, msg string) {
			s.storage.Log().Write(msg)
		},
	}
	s.thread.SetLocal("storage", s.storage)
	return s, nil
}

func (s *SStarLark) Run(ctx context.Context) (code int64, err error) {
	s.thread.SetLocal("star_ctx", ctx)
	defer func() {
		if _r := recover(); _r != nil {
			err = fmt.Errorf("panic during execution %v", _r)
			code = common.CodeSystemErr
			stack := debug.Stack()
			if _err, ok := _r.(error); ok && strings.Contains(_err.Error(), context.Canceled.Error()) {
				code = common.CodeKilled
				err = common.ErrManual
			}
			s.storage.Log().Write(err.Error(), string(stack))
		}
	}()

	content, err := s.storage.Content()
	if err != nil {
		return common.CodeFailed, err
	}

	var predeclared = star_libs.StarlarkPredeclared
	predeclared["workspace"] = starlark.String(s.workspace)
	predeclared["log"] = s.logModule()
	evalFnVal, err := starlark.ExecFileOptions(&syntax.FileOptions{
		Set:             true,
		While:           true,
		TopLevelControl: true,
		GlobalReassign:  true,
		Recursion:       true,
	}, s.thread, s.name+".star", content, predeclared)
	if err != nil {
		return common.CodeFailed, err
	}
	evalFn, ok := evalFnVal["EvalCall"]
	if !ok {
		return common.CodeFailed, errors.New("not found EvalCall")
	}

	_, err = starlark.Call(s.thread, evalFn, starlark.Tuple{
		s.getParam(),
	}, nil)
	return common.CodeSuccess, nil
}

func (s *SStarLark) getParam() *starlark.Dict {
	var paramDict = new(starlark.Dict)
	taskEnv := s.storage.GlobalEnv().List()
	for _, v := range taskEnv {
		_ = paramDict.SetKey(starlark.String(v.Name), starlark.String(v.Value))
	}
	stepEnv := s.storage.Env().List()
	for _, v := range stepEnv {
		_ = paramDict.SetKey(starlark.String(v.Name), starlark.String(v.Value))
	}
	return paramDict
}
func (s *SStarLark) Clear() error {
	s.thread.Cancel("user interrupt")
	return nil
}
