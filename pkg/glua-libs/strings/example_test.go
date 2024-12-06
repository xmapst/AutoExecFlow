package strings

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

// strings.split(string, sep)
func Test_Split(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local inspect = require("inspect")
    local strings = require("strings")
    local result = strings.split("a b c d", " ")
    print(inspect(result, {newline="", indent=""}))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// { "a", "b", "c", "d" }
}

// strings.fields(string)
func Test_Fields(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local strings = require("strings")
    local inspect = require("inspect")
    local result = strings.fields("a b c d")
    print(inspect(result, {newline="", indent=""}))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// { "a", "b", "c", "d" }
}

// strings.has_prefix(string, prefix)
func Test_HasPrefix(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local strings = require("strings")
    local result = strings.has_prefix("abcd", "a")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// true
}

// strings.has_suffix(string, suffix)
func Test_HasSuffix(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local strings = require("strings")
    local result = strings.has_suffix("abcd", "d")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// true
}

// strings.trim(string, cutset)
func Test_Trim(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local strings = require("strings")
    local result = strings.trim("abcd", "d")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// abc
}

// strings.trim_prefix(string, cutset)
func Test_TrimPrefix(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local strings = require("strings")
    local result = strings.trim_prefix("abcd", "d")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// abcd
}

// strings.trim_suffix(string, cutset)
func Test_TrimSuffix(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local strings = require("strings")
    local result = strings.trim_suffix("abcd", "d")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// abc
}

// strings.contains(string, substring)
func Test_Contains(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local strings = require("strings")
    local result = strings.contains("abcd", "d")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// true
}
