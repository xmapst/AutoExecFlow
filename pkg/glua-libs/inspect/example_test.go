package inspect

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// inspect(obj)
func Test_full(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
	local inspect = require("inspect")
    local table = {a={b=2}}
    print(inspect(table, {newline="", indent=""}))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	//{a = {b = 2}}
}
