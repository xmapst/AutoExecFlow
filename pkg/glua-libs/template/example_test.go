package template

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

func Test_package(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
local template = require("template")

local mustache, err = template.choose("mustache")

local values = {name="world"}
print( mustache:render("Hello {{name}}!", values) ) -- mustache:render_file(filename values)

local values = {data = {"one", "two"}}
print( mustache:render("{{#data}} {{.}} {{/data}}", values) )
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// Hello world!
	//  one  two
}
