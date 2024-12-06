package socketcore_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestMasterBind(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	script := fmt.Sprintf(`require 'socket'.bind('%s', '%s')`, "localhost", "8383")
	assertions.NoError(L.DoString(script))
}
