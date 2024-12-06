package shellescape

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func Test_Quote(t *testing.T) {
	L := lua.NewState()
	Preload(L)
	source := `
local shellescape = require("shellescape")
print(shellescape.quote("foo"))
`
	if err := L.DoString(source); err != nil {
		t.Fatal(err)
	}
	// Output:
	// foo
}

func Test_QuoteCommand(t *testing.T) {
	L := lua.NewState()
	Preload(L)
	source := `
local shellescape = require("shellescape")
print(shellescape.quote_command({"echo", "foo bar baz"}))
`
	if err := L.DoString(source); err != nil {
		t.Fatal(err)
	}
	// Output:
	// echo 'foo bar baz'
}

func Test_StripUnsafe(t *testing.T) {
	L := lua.NewState()
	Preload(L)
	source := `
local shellescape = require("shellescape")
print(shellescape.strip_unsafe("foo\nbar"))
`
	if err := L.DoString(source); err != nil {
		t.Fatal(err)
	}
	// Output:
	// foobar
}
