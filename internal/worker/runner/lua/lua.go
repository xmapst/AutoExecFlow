package lua

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	glualibs "github.com/xmapst/AutoExecFlow/pkg/glua-libs"
)

type SLua struct {
	lua       *lua.LState
	storage   storage.IStep
	workspace string
}

func New(storage storage.IStep, workspace string) (*SLua, error) {
	l := &SLua{
		storage:   storage,
		workspace: workspace,
	}
	l.newLuaState()
	return l, nil
}

func (l *SLua) Clear() error {
	if l.lua.IsClosed() {
		return nil
	}
	l.lua.Close()
	return nil
}

func (l *SLua) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		if _r := recover(); _r != nil {
			err = fmt.Errorf("panic during execution %v", _r)
			code = common.CodeSystemErr
			stack := debug.Stack()
			if _err, ok := _r.(error); ok && strings.Contains(_err.Error(), context.Canceled.Error()) {
				code = common.CodeKilled
				err = common.ErrManual
			}
			l.storage.Log().Write(err.Error(), string(stack))
		}
	}()

	content, err := l.storage.Content()
	if err != nil {
		return common.CodeSystemErr, err
	}

	l.lua.SetContext(ctx)
	var params = map[string]any{}
	taskEnv := l.storage.GlobalEnv().List()
	for _, v := range taskEnv {
		params[v.Name] = v.Value
	}
	stepEnv := l.storage.Env().List()
	for _, v := range stepEnv {
		params[v.Name] = v.Value
	}
	l.lua.SetGlobal("workspace", luar.New(l.lua, l.workspace))
	err = l.lua.DoString(content)
	if err != nil {
		return common.CodeFailed, err
	}
	evalFn := l.lua.GetGlobal("EvalCall")
	if evalFn.Type() != lua.LTFunction {
		return common.CodeFailed, errors.New("EvalCall function not found")
	}
	if err = l.lua.CallByParam(lua.P{
		Fn: evalFn,
	}, luar.New(l.lua, params)); err != nil {
		return common.CodeFailed, err
	}
	return common.CodeSuccess, nil
}

func (l *SLua) newLuaState() {
	l.lua = lua.NewState(lua.Options{IncludeGoStackTrace: true})

	// load module
	glualibs.Preload(l.lua)

	// 默认通用Global符号
	l.lua.SetGlobal("debugf", luar.New(l.lua, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[DEBUG] %s", format), args...)
	}))
	l.lua.SetGlobal("debug", luar.New(l.lua, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[DEBUG] %s", args...))
	}))
	l.lua.SetGlobal("infof", luar.New(l.lua, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[INFO] %s", format), args...)
	}))
	l.lua.SetGlobal("info", luar.New(l.lua, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[INFO] %s", args...))
	}))
	l.lua.SetGlobal("warnf", luar.New(l.lua, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[WARN] %s", format), args...)
	}))
	l.lua.SetGlobal("warn", luar.New(l.lua, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[WARN] %s", args...))
	}))
	l.lua.SetGlobal("errorf", luar.New(l.lua, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[ERROR] %s", format), args...)
	}))
	l.lua.SetGlobal("error", luar.New(l.lua, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[ERROR] %s", args...))
	}))
	l.lua.SetGlobal("printf", luar.New(l.lua, l.outPrintf))
	l.lua.SetGlobal("print", luar.New(l.lua, l.outPrint))

}

func (l *SLua) outPrintf(format string, args ...any) {
	l.storage.Log().Writef(format, args...)
}

func (l *SLua) outPrint(args ...any) {
	l.storage.Log().Write(fmt.Sprint(args...))
}
