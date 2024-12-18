package yaml

import (
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/io"
)

const (
	yamlEncoderType = "yaml.Encoder"
)

func CheckYAMLEncoder(L *lua.LState, n int) *yaml.Encoder {
	ud := L.CheckUserData(n)
	if encoder, ok := ud.Value.(*yaml.Encoder); ok {
		return encoder
	}
	L.ArgError(n, yamlEncoderType+" expected")
	return nil
}

func LVYAMLEncoder(L *lua.LState, encoder *yaml.Encoder) lua.LValue {
	ud := L.NewUserData()
	ud.Value = encoder
	L.SetMetatable(ud, L.GetTypeMetatable(yamlEncoderType))
	return ud
}

func yamlEncoderEncode(L *lua.LState) int {
	encoder := CheckYAMLEncoder(L, 1)
	arg := L.CheckAny(2)
	if err := encoder.Encode(marshalValue{
		LValue:  arg,
		visited: make(map[*lua.LTable]bool),
	}); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func registerYAMLEncoder(L *lua.LState) {
	mt := L.NewTypeMetatable(yamlEncoderType)
	L.SetGlobal(yamlEncoderType, mt)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"encode": yamlEncoderEncode,
	}))
}

func newYAMLEncoder(L *lua.LState) int {
	writer := io.CheckIOWriter(L, 1)
	L.Pop(L.GetTop())
	encoder := yaml.NewEncoder(writer)
	L.Push(LVYAMLEncoder(L, encoder))
	return 1
}
