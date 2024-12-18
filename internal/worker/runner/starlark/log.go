package starlark

import (
	"fmt"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func (s *SStarLark) logModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: "log",
		Members: starlark.StringDict{
			"debug": starlark.NewBuiltin("debug", s.logDebug),
			"info":  starlark.NewBuiltin("info", s.logInfo),
			"warn":  starlark.NewBuiltin("warn", s.logWarn),
			"error": starlark.NewBuiltin("error", s.logError),
		},
	}
}

func (s *SStarLark) logDebug(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return s.createLogFunc(thread, b, args, kwargs, "debug")
}

func (s *SStarLark) logInfo(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return s.createLogFunc(thread, b, args, kwargs, "info")
}

func (s *SStarLark) logWarn(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return s.createLogFunc(thread, b, args, kwargs, "warn")
}

func (s *SStarLark) logError(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return s.createLogFunc(thread, b, args, kwargs, "error")
}

func (s *SStarLark) createLogFunc(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
	level string,
) (starlark.Value, error) {
	timeStr := time.Now().Local().Format("2006-01-02 15:04:05")
	var starCall = thread.CallFrame(1)
	var msgs interface{}
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "*", &msgs); err != nil {
		return nil, err
	}
	s.storage.Log().Write(timeStr, level, starCall.Pos.String(), fmt.Sprintf("%v", msgs))
	return starlark.None, nil
}
