package sockethttp_test

import (
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket/sockethttp"
)

func TestWebHdfsUrlScheme(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()

	// Start an HTTP endpoint for LuaSocket to connect to
	listener, err := net.Listen("tcp", "localhost:0")
	assertions.NoError(err)
	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		_ = http.Serve(listener, nil)
	}()

	// Make a webdav request
	L.PreloadModule("socket.http", sockethttp.Loader)
	url := fmt.Sprintf("webhdfs://localhost:%d", port)
	assertions.NoError(L.DoString(fmt.Sprintf(`require 'socket.http'.request('%s')`, url)))
}
