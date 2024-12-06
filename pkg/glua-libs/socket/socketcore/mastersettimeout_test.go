package socketcore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestMasterSetTimeout(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	assertions.NoError(L.DoString(`return require 'socket.core'.tcp():settimeout(.25)`))
	retval := L.Get(-1)
	assertions.Equal(lua.LTNumber, retval.Type())
}
