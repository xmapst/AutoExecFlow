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

func TestConnect(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	listener, err := net.Listen("tcp", "localhost:0")
	assertions.NoError(err)
	port := listener.Addr().(*net.TCPAddr).Port

	accepted := false
	go func() {
		if _, err := listener.Accept(); err == nil {
			accepted = true
		}
	}()

	script := fmt.Sprintf(`c=require 'socket.core'.connect('%s', %d); c:close()`, "127.0.0.1", port)
	assertions.NoError(L.DoString(script))

	time.Sleep(20 * time.Millisecond)
	assertions.True(accepted)

	assertions.NoError(listener.Close())
}
