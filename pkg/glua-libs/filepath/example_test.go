//go:build !windows && !plan9

package filepath

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

// filepath.ext(string)
func Test_Ext(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local filepath = require("filepath")
    local result = filepath.ext("/var/tmp/file.name")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// .name
}

// filepath.basename(string)
func Test_Basename(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local filepath = require("filepath")
    local result = filepath.basename("/var/tmp/file.name")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// file.name
}

// filepath.basename(string)
func Test_Dir(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local filepath = require("filepath")
    local result = filepath.dir("/var/tmp/file.name")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// /var/tmp
}

// filepath.basename(string)
func Test_Join(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local filepath = require("filepath")
    local result = filepath.join("var", "tmp", "file.name")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// var/tmp/file.name
}

// filepath.glob(string)
func Test_Glob(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local filepath = require("filepath")
    local inspect = require("inspect")
    local result = filepath.glob("./*/*.lua")
    print(inspect(result, {newline="", indent=""}))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// { "test/test_api.lua" }
}
