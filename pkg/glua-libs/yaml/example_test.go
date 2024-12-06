package yaml

import (
	"log"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

// yaml.decode(string)
func Example() {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local yaml = require("yaml")
    local inspect = require("inspect")
    local text = [[
a:
  b: 1
    ]]
    local result, err = yaml.decode(text)
    if err then error(err) end
    print(inspect(result, {newline="", indent=""}))
`
	if err := state.DoString(source); err != nil {
		log.Fatal(err.Error())
	}
	// Output:
	// {a = {b = 1}}
}

func ExampleEncode() {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local yaml = require("yaml")
    local encoded, err = yaml.encode({a = {b = 1}})
    if err then error(err) end
    print(encoded)
`
	if err := state.DoString(source); err != nil {
		log.Fatal(err.Error())
	}
	// Output:
	// a:
	//   b: 1
	//
}
