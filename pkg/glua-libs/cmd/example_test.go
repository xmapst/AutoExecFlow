package cmd

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/runtime"
)

// cmd.exec()
func Test_Exec(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	runtime.Preload(state)
	source := `
local cmd = require("cmd")
local runtime = require("runtime")

local command = "sleep 1"
if runtime.goos() == "windows" then command = "timeout 1" end

local result, err = cmd.exec(command)
if err then error(err) end
print(result.status)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// 0
}
