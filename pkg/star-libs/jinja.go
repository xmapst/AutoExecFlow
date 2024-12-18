package star_libs

import (
	"fmt"
	"os"

	json "github.com/json-iterator/go"
	"github.com/qri-io/starlib/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/xmapst/AutoExecFlow/pkg/jinja"
)

var jinjaModule = &starlarkstruct.Module{
	Name: "jinja",
	Members: starlark.StringDict{
		"parse":      starlark.NewBuiltin("parse", jinjaParse),
		"parse_file": starlark.NewBuiltin("parse_file", jinjaParseFile),
	},
}

func jinjaParse(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var template string
	var data starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "template", &template, "data?", &data); err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	_data, err := convertToMap(data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	res, err := jinja.Parse(template, _data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	return starlark.Tuple{
		starlark.String(res),
		starlark.None,
	}, nil
}

func jinjaParseFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var filename string
	var data starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "filename", &filename, "data?", &data); err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	template, err := os.ReadFile(filename)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	_data, err := convertToMap(data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	res, err := jinja.Parse(string(template), _data)
	if err != nil {
		return starlark.Tuple{
			starlark.None,
			starlark.String(err.Error()),
		}, err
	}

	return starlark.Tuple{
		starlark.String(res),
		starlark.None,
	}, nil
}

func convertToMap(value starlark.Value) (map[string]any, error) {
	switch value.(type) {
	case starlark.NoneType:
		value = starlark.String("{}")
	case starlark.String, starlark.Mapping:
	default:
		return nil, fmt.Errorf("got %s, want string or mapping", value.Type())
	}

	data, err := util.Unmarshal(value)
	if err != nil {
		return nil, err
	}
	bs, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var dataM map[string]any
	err = json.Unmarshal(bs, &dataM)
	return dataM, err
}
