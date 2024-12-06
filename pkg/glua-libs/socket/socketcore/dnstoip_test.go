package socketcore_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestDnsToIpLocalhost(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	assertions.NoError(L.DoString("return require 'socket.core'.dns.toip('localhost')"))
	assertions.Equal("127.0.0.1", L.ToString(-1))
}
