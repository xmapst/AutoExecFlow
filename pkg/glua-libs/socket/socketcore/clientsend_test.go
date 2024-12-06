package socketcore_test

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestClientSend(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	listener, err := net.Listen("tcp", "localhost:0")
	assertions.NoError(err)
	port := listener.Addr().(*net.TCPAddr).Port

	accepted := false
	closed := false
	received := make([]string, 0)
	go func() {
		if conn, err := listener.Accept(); err == nil {
			accepted = true
			for {
				buffer := make([]byte, 1000)
				bytesRead, err := conn.Read(buffer)
				if err == io.EOF {
					closed = true
					break
				} else {
					str := string(buffer[:bytesRead])
					received = append(received, str)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	script := fmt.Sprintf(`c=require 'socket.core'.connect('%s', %d); c:send('abc'); c:close()`, "127.0.0.1", port)
	assertions.NoError(L.DoString(script))

	time.Sleep(20 * time.Millisecond)
	assertions.True(accepted)
	assertions.True(closed)
	assertions.Equal(1, len(received))
	assertions.Equal("abc", received[0])
	assertions.NoError(listener.Close())
}

func TestClientSendWithSubstring(t *testing.T) {
	assertions := assert.New(t)
	L := lua.NewState()
	defer L.Close()
	socket.Preload(L)

	listener, err := net.Listen("tcp", "localhost:0")
	assertions.NoError(err)
	port := listener.Addr().(*net.TCPAddr).Port

	accepted := false
	closed := false
	received := make([]string, 0)
	go func() {
		if conn, err := listener.Accept(); err == nil {
			accepted = true
			for {
				buffer := make([]byte, 1000)
				bytesRead, err := conn.Read(buffer)
				if err == io.EOF {
					closed = true
					break
				} else {
					str := string(buffer[:bytesRead])
					received = append(received, str)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	script := fmt.Sprintf(`c=require 'socket.core'.connect('%s', %d); x=c:send('abcde', 4); c:close(); return x`, "127.0.0.1", port)
	assertions.NoError(L.DoString(script))

	time.Sleep(20 * time.Millisecond)
	assertions.True(accepted)
	assertions.True(closed)
	assertions.Equal(2, L.ToInt(-1)) // return value x indicates 2 bytes written
	assertions.Equal(1, len(received))
	assertions.Equal("de", received[0])
	assertions.NoError(listener.Close())
}
