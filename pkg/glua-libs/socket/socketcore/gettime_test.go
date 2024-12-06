package socketcore_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket/socketcore"
)

func TestGettime(t *testing.T) {
	assertions := assert.New(t)

	luaState := lua.NewState()
	defer luaState.Close()

	luaState.PreloadModule("socket.core", socketcore.Loader)

	now := time.Now()
	assertions.NoError(luaState.DoString("return require 'socket.core'.gettime()"))

	lv := luaState.Get(-1)
	retval, ok := lv.(lua.LNumber)

	assertions.True(ok)
	expectedMin := float64(now.UnixNano()) / 1e9
	assertions.True(float64(retval) >= expectedMin)
}
