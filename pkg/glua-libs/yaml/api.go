// Package yaml implements yaml decode functionality for lua.
package yaml

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"
	luar "layeh.com/gopher-luar"
)

// Decode lua yaml.decode(string) returns (table, error)
func Decode(L *lua.LState) int {
	str := L.CheckString(1)

	var value interface{}
	err := yaml.Unmarshal([]byte(str), &value)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(luar.New(L, value))
	return 1
}

// Encode lua yaml.encode(any) returns (string, error)
func Encode(L *lua.LState) int {
	arg := L.CheckAny(1)
	data, err := yaml.Marshal(marshalValue{
		LValue:  arg,
		visited: make(map[*lua.LTable]bool),
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(data))
	return 1
}

func tableIsSlice(table *lua.LTable) bool {
	expectedKey := lua.LNumber(1)
	for key, _ := table.Next(lua.LNil); key != lua.LNil; key, _ = table.Next(key) {
		if expectedKey != key {
			return false
		}
		expectedKey++
	}
	return true
}

var _ yaml.Marshaler = marshalValue{}

type marshalValue struct {
	lua.LValue
	visited map[*lua.LTable]bool
}

type null struct {
	lua.LValue
}

func (null) String() string       { return "null" }
func (null) Type() lua.LValueType { return lua.LTNil }

var _ lua.LValue = null{}

var (
	errNested      = errors.New("cannot encode recursively nested tables to YAML")
	errSparseArray = errors.New("cannot encode sparse array")
	errInvalidKeys = errors.New("cannot encode mixed or invalid key types")
)

func (j marshalValue) MarshalYAML() (interface{}, error) {
	var data interface{}

	switch converted := j.LValue.(type) {
	case lua.LBool:
		data = bool(converted)
	case lua.LNumber:
		data = float64(converted)
	case *lua.LNilType:
		data = nil
	case null:
		data = nil
	case lua.LString:
		data = string(converted)
	case *lua.LTable:
		if j.visited[converted] {
			return nil, errNested
		}
		j.visited[converted] = true

		key, value := converted.Next(lua.LNil)

		switch key.Type() {
		case lua.LTNil: // empty table
			data = []int{}
		case lua.LTNumber:
			arr := make([]marshalValue, 0, converted.Len())
			expectedKey := lua.LNumber(1)
			for key != lua.LNil {
				if key.Type() != lua.LTNumber {
					return nil, errInvalidKeys
				}
				if expectedKey != key {
					return nil, errSparseArray
				}
				arr = append(arr, marshalValue{value, j.visited})
				expectedKey++
				key, value = converted.Next(key)
			}
			data = arr
		case lua.LTString:
			obj := make(map[string]marshalValue)
			for key != lua.LNil {
				if key.Type() != lua.LTString {
					return nil, errInvalidKeys
				}
				obj[key.String()] = marshalValue{value, j.visited}
				key, value = converted.Next(key)
			}
			data = obj
		default:
			return nil, errInvalidKeys
		}
	default:
		return nil, fmt.Errorf("cannot encode %s to YAML", j.LValue.Type().String())
	}
	return data, nil
}
