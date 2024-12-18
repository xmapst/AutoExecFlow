package lua

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

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
	fn, err := l.lua.Load(strings.NewReader(content), "main.lua")
	if err != nil {
		return common.CodeFailed, err
	}
	l.lua.Push(fn)
	if err = l.lua.PCall(0, lua.MultRet, nil); err != nil {
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
	l.lua.PreloadModule("log", l.loggerLoader)

	// 默认通用Global符号
	l.lua.SetGlobal("printf", luar.New(l.lua, l.outPrintf))
	l.lua.SetGlobal("print", luar.New(l.lua, l.outPrint))

}

func (l *SLua) outPrintf(format string, args ...any) {
	l.storage.Log().Writef(format, args...)
}

func (l *SLua) outPrint(args ...any) {
	l.storage.Log().Write(fmt.Sprint(args...))
}

func (l *SLua) loggerLoader(L *lua.LState) int {
	t := L.NewTable()
	L.SetFuncs(t, map[string]lua.LGFunction{
		"debug":  l.createLogFunc("DEBUG", false),
		"info":   l.createLogFunc("INFO", false),
		"warn":   l.createLogFunc("WARN", false),
		"error":  l.createLogFunc("ERROR", false),
		"debugf": l.createLogFunc("DEBUG", true),
		"infof":  l.createLogFunc("INFO", true),
		"warnf":  l.createLogFunc("WARN", true),
		"errorf": l.createLogFunc("ERROR", true),
	})
	L.Push(t)
	return 1
}

func (l *SLua) createLogFunc(level string, isFormatted bool) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		var args []interface{}
		var logMessage string

		// Collect arguments based on whether it's a formatted log
		if isFormatted {
			format := L.CheckString(1)
			for i := 2; i <= L.GetTop(); i++ {
				args = append(args, L.Get(i))
			}
			logMessage = fmt.Sprintf(format, args...)
		} else {
			for i := 1; i <= L.GetTop(); i++ {
				args = append(args, L.Get(i))
			}
			logMessage = fmt.Sprint(args...)
		}

		// Add timestamp and log level
		timeStr := time.Now().Local().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("%s %s %s%s", timeStr, level, L.Where(1), logMessage)

		// Call the appropriate logging function
		l.outPrint(logLine)
		return 0
	}
}
