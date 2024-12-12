package lua

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	glualibs "github.com/xmapst/AutoExecFlow/pkg/glua-libs"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
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
			stack := debug.Stack()
			code = common.CodeSystemErr
			if err != nil {
				err = fmt.Errorf("panic during execution %v %v", err, _r)
				return
			}
			err = fmt.Errorf("panic during execution %v %s", _r, stack)
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
	var done = make(chan error, 1)
	defer close(done)
	go func() {
		defer func() {
			if _r := recover(); _r != nil {
				stack := debug.Stack()
				logx.Errorln(_r, string(stack))
				done <- errors.New("panic during execution")
			}
		}()

		_err := l.lua.DoString(content)
		if err != nil {
			done <- _err
			return
		}
		evalFn := l.lua.GetGlobal("EvalCall")
		if evalFn.Type() != lua.LTFunction {
			done <- errors.New("EvalCall function not found")
			return
		}
		if _err = l.lua.CallByParam(lua.P{
			Fn: evalFn,
		}, luar.New(l.lua, params)); _err != nil {
			done <- _err
			return
		}
		done <- nil
	}()

	select {
	case err = <-done:
		if err != nil {
			l.storage.Log().Writef("[ERROR] %v", err)
			return common.CodeFailed, err
		}
		return common.CodeSuccess, nil
	case <-ctx.Done():
		l.lua.Close()
		return common.CodeFailed, errors.New("has been killed")
	}
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
