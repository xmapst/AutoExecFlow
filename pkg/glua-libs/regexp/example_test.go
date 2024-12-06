package regexp

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

// regexp_ud:match(string)
func Test_Match(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
    local regexp = require("regexp")
    local reg, err = regexp.compile("hello")
    if err then error(err) end
    local result = reg:match("string: 'hello world'")
    print(result)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// true
}

// regexp_ud:find_all_string_submatch(string)
func Test_FindAllStringSubmatch(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local regexp = require("regexp")
    local inspect = require("inspect")
    local reg, err = regexp.compile("string: '(.*)\\s+(.*)'$")
    if err then error(err) end
    local result = reg:find_all_string_submatch("string: 'hello world'")
    print(inspect(result, {newline="", indent=""}))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// { { "string: 'hello world'", "hello", "world" } }
}
