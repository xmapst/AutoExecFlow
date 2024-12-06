package socketcore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestSkip0(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	assertions.NoError(L.DoString(`return require 'socket.core'.skip(0, 1, 2, 3)`))
	assertions.Equal(3, L.GetTop())
	assertions.Equal(1, L.ToInt(-3))
	assertions.Equal(2, L.ToInt(-2))
	assertions.Equal(3, L.ToInt(-1))
}

func TestSkip1(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	assertions.NoError(L.DoString(`return require 'socket.core'.skip(1, 1, 2, 3)`))
	assertions.Equal(2, L.GetTop())
	assertions.Equal(2, L.ToInt(-2))
	assertions.Equal(3, L.ToInt(-1))
}

func TestSkip2(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	assertions.NoError(L.DoString(`return require 'socket.core'.skip(2, 1, 2, 3)`))
	assertions.Equal(1, L.GetTop())
	assertions.Equal(3, L.ToInt(-1))
}

func TestSkipPast(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	assertions.NoError(L.DoString(`return require 'socket.core'.skip(9, 1, 2, 3)`))
	assertions.Equal(0, L.GetTop())
}
