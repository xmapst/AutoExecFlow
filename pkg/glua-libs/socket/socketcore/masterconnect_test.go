package socketcore_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestMasterConnect(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	listener, err := net.Listen("tcp", "localhost:0")
	assertions.NoError(err)
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port

	accepted := false
	go func() {
		if _, err := listener.Accept(); err == nil {
			accepted = true
		}
	}()

	script := fmt.Sprintf(`require 'socket.core'.tcp():connect('%s', %d)`, "127.0.0.1", port)
	assertions.NoError(L.DoString(script))

	time.Sleep(20 * time.Millisecond)
	assertions.True(accepted)
}
