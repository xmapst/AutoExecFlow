package star_libs

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/go-cmd/cmd"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"

	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

// StarlarkPredeclared builtins is a Starlark module of Python-like builtins functions.
var StarlarkPredeclared = starlark.StringDict{}

func init() {
	assertModule, err := starlarktest.LoadAssertModule()
	if err != nil {
		logx.Fatalln(err)
	}
	StarlarkPredeclared = starlark.StringDict{
		"assert":       assertModule["assert"],
		"time":         timeModule,
		"jinja":        jinjaModule,
		"resty":        restyModule(),
		"read_file":    starlark.NewBuiltin("read_file", readFile),
		"write_file":   starlark.NewBuiltin("write_file", writeFile),
		"exec_command": starlark.NewBuiltin("exec_command", execCommand),
		"map":          starlark.NewBuiltin("map", map_),
		"next":         starlark.NewBuiltin("next", next),
		"filter":       starlark.NewBuiltin("filter", filter),
		"callable":     starlark.NewBuiltin("callable", callable),
		"hex":          starlark.NewBuiltin("hex", hex),
		"oct":          starlark.NewBuiltin("oct", oct),
		"bin":          starlark.NewBuiltin("bin", bin),
	}
}

func readFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var filename string
	if err := starlark.UnpackArgs("read_file", args, kwargs, "filename", &filename); err != nil {
		return nil, err
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read_file: %v", err)
	}
	return starlark.String(content), nil
}

func writeFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		filename string
		content  string
	)
	if err := starlark.UnpackArgs("write_file", args, kwargs, "filename", &filename, "content", &content); err != nil {
		return nil, err
	}
	return starlark.None, os.WriteFile(filename, []byte(content), os.ModePerm)
}
func execCommand(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	defer func() {
		if _r := recover(); _r != nil {
			logx.Errorln("panic during execution", _r, string(debug.Stack()))
			return
		}
	}()

	ctx := thread.Local("star_ctx").(context.Context)
	var command string
	if err := starlark.UnpackArgs("exec_command", args, kwargs, "cmd", &command); err != nil {
		return starlark.Tuple{
			starlark.MakeInt(255),
			starlark.String(""),
			starlark.String(err.Error()),
		}, err
	}

	shell := cmd.NewCmd("sh", "-c", command)
	if runtime.GOOS == "windows" {
		shell = cmd.NewCmd("cmd", "/C", command)
	}
	defer func(shell *cmd.Cmd) {
		_ = shell.Stop()
	}(shell)

	shell.Dir = os.TempDir()
	shell.Env = os.Environ()

	select {
	case <-ctx.Done():
		_ = shell.Stop()
		return starlark.Tuple{
			starlark.MakeInt(255),
			starlark.String(""),
			starlark.String("context canceled"),
		}, nil
	case status := <-shell.Start():
		return starlark.Tuple{
			starlark.MakeInt(status.Exit),
			starlark.String(strings.Join(status.Stdout, "\n")),
			starlark.String(strings.Join(status.Stderr, "\n")),
		}, nil
	}
}

func formatInt(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
	f string,
) (starlark.Value, error) {
	var i starlark.Int
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &i); err != nil {
		return nil, err
	}

	if v, ok := i.Int64(); ok {
		return starlark.String(fmt.Sprintf(f, v)), nil
	}

	return starlark.String(fmt.Sprintf(f, i.BigInt())), nil
}

func bin(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return formatInt(thread, b, args, kwargs, "%#b")
}

func oct(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return formatInt(thread, b, args, kwargs, "%O")
}

func hex(
	thread *starlark.Thread,
	b *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return formatInt(thread, b, args, kwargs, "%#x")
}

// asString unquotes a starlark string value
func asString(x starlark.Value) (string, error) {
	return strconv.Unquote(x.String())
}

func asStringSlice(x starlark.Value) ([]string, error) {
	seq, ok := x.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("expected a sequence, got %s", x.Type())
	}

	var result []string
	it := seq.Iterate()
	defer it.Done()

	var v starlark.Value
	for it.Next(&v) {
		str, ok := v.(starlark.String)
		if !ok {
			return nil, fmt.Errorf("expected string elements, got %s", v.Type())
		}
		result = append(result, str.GoString())
	}
	return result, nil
}

func convertDictToMap(dict *starlark.Dict) (map[string]string, error) {
	keys := dict.Keys()
	if len(keys) == 0 {
		return nil, nil
	}
	var result = make(map[string]string)
	for _, key := range keys {
		keystr, err := asString(key)
		if err != nil {
			return nil, err
		}

		val, _, err := dict.Get(key)
		if err != nil {
			return nil, err
		}
		if val.Type() != "string" {
			return nil, fmt.Errorf("expected param value for key '%s' to be a string. got: '%s'", key, val.Type())
		}
		valstr, err := asString(val)
		if err != nil {
			return nil, err
		}
		result[keystr] = valstr
	}

	return result, nil
}
