package socketcore_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket/socketcore"
)

func TestDnsGetHostName(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("socket.core", socketcore.Loader)

	expected, err := os.Hostname()
	assertions.NoError(err)

	assertions.NoError(L.DoString(`return require 'socket.core'.dns.gethostname()`))
	actual := L.Get(-1)

	assertions.Equal(expected, actual.String())
}
