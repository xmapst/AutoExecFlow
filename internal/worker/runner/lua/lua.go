package lua

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	glualibs "github.com/xmapst/AutoExecFlow/pkg/glua-libs"
)

type SLua struct {
	storage   storage.IStep
	workspace string
}

func New(storage storage.IStep, workspace string) (*SLua, error) {
	l := &SLua{
		storage:   storage,
		workspace: workspace,
	}
	return l, nil
}

func (l *SLua) Clear() error {
	return nil
}

func (l *SLua) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		if _r := recover(); _r != nil {
			code = common.CodeSystemErr
			if err != nil {
				err = fmt.Errorf("panic during execution %v %v", err, _r)
				return
			}
			err = fmt.Errorf("panic during execution %v", _r)
		}
	}()

	content, err := l.storage.Content()
	if err != nil {
		return common.CodeSystemErr, err
	}

	L := l.newLuaState(nil)
	defer L.Close()

	L.SetContext(ctx)

	var params = map[string]any{}
	taskEnv := l.storage.GlobalEnv().List()
	for _, v := range taskEnv {
		params[v.Name] = v.Value
	}
	stepEnv := l.storage.Env().List()
	for _, v := range stepEnv {
		params[v.Name] = v.Value
	}
	L.SetGlobal("workspace", luar.New(L, l.workspace))

	err = L.DoString(content)
	if err != nil {
		return common.CodeSystemErr, err
	}
	evalFn := L.GetGlobal("EvalCall")
	if evalFn.Type() != lua.LTFunction {
		return common.CodeSystemErr, errors.New("EvalCall function not found")
	}
	if err = L.CallByParam(lua.P{
		Fn: evalFn,
	}, luar.New(L, params)); err != nil {
		return common.CodeSystemErr, err
	}
	return common.CodeSuccess, nil
}

func (l *SLua) newLuaState(globals map[string]interface{}) *lua.LState {
	L := lua.NewState(lua.Options{IncludeGoStackTrace: true})

	// load module
	glualibs.Preload(L)

	// 默认通用Global符号
	L.SetGlobal("debugf", luar.New(L, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[DEBUG] %s", format), args...)
	}))
	L.SetGlobal("debug", luar.New(L, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[DEBUG] %s", args...))
	}))
	L.SetGlobal("infof", luar.New(L, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[INFO] %s", format), args...)
	}))
	L.SetGlobal("info", luar.New(L, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[INFO] %s", args...))
	}))
	L.SetGlobal("warnf", luar.New(L, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[WARN] %s", format), args...)
	}))
	L.SetGlobal("warn", luar.New(L, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[WARN] %s", args...))
	}))
	L.SetGlobal("errorf", luar.New(L, func(format string, args ...any) {
		l.outPrintf(fmt.Sprintf("[ERROR] %s", format), args...)
	}))
	L.SetGlobal("error", luar.New(L, func(args ...any) {
		l.outPrintf(fmt.Sprintf("[ERROR] %s", args...))
	}))
	L.SetGlobal("printf", luar.New(L, l.outPrintf))
	L.SetGlobal("print", luar.New(L, l.outPrint))

	for k, v := range globals {
		L.SetGlobal(k, luar.New(L, v))
	}

	return L
}

func (l *SLua) outPrintf(format string, args ...any) {
	l.storage.Log().Writef(format, args...)
}

func (l *SLua) outPrint(args ...any) {
	l.storage.Log().Write(fmt.Sprint(args...))
}
