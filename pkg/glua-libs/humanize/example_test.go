package humanize

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

// humanize.ibytes(number)
func Test_IBytes(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
local humanize = require("humanize")
print(humanize.ibytes(1395864371))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// 1.3 GiB
}

// humanize.parse_bytes(string)
func Test_ParseBytes(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
local humanize = require("humanize")
print(humanize.parse_bytes("1.3GiB"))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// 1395864371
}

// humanize.time(number)
func Test_Time(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	time.Preload(state)
	source := `
local humanize = require("humanize")
local time = require("time")
print(humanize.time(time.unix() - 61))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// 1 minute ago
}

// humanize.si(input, unit)
func Test_SI(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	source := `
local humanize = require("humanize")
print(humanize.si(0.212121, "m"))
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// 212.121 mm
}
