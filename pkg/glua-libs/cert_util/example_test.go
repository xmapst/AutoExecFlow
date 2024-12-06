package cert_util

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

// cert_util.not_after("host", <ip:port>)
func Test_package(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	time.Preload(state)
	source := `
    local cert_util = require("cert_util")
    local time = require("time")
    local tx, err = cert_util.not_after("google.com", "64.233.165.101:443")
    if err then error(err) end
    print(tx > time.unix())
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// true
}
