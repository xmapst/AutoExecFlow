package log

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// print(args..)
func Test_Print(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local log = require("log")
    local info = log.new("STDOUT", "[INFO] ")
    info:print("1 ", 2)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// [INFO] 1 2
}

// printf(string, args..)
func Test_Printf(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local log = require("log")
    local info = log.new("STDOUT", "[INFO] ")
    info:printf("%s %d\n", "1", 2)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// [INFO] 1 2
}

// println(string, args..)
func Test_Println(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local log = require("log")
    local info = log.new("STDOUT", "[INFO] ")
    info:println("1", 2)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// [INFO] 1 2
}

// set_flags(config={})
func Test_SetFlags(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local log = require("log")
    local logger = log.new()
    logger:set_prefix("[prefix] ")
    logger:set_flags({longfile=true})
    logger:println("1", 2)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// [prefix] <string>:6: 1 2
}
