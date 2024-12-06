// build +linux +amd64

package runtime

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

// runtime.goos(), runtime.goarch()
func Test_package(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	inspect.Preload(state)
	source := `
    local runtime = require("runtime")
    print(runtime.goos())
    print(runtime.goarch())
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// linux
	// amd64
}
