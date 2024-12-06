package socket_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/socket"
)

func TestExcept(t *testing.T) {
	doTest("excepttest.lua", t)
}

//func TestFtp(t *testing.T) {
//	doTest("ftptest.lua", t)
//}

//func TestHttp(t *testing.T) {
//	doTest("httptest.lua", t)
//}

//func TestLtn12(t *testing.T) {
//	doTest("ltn12test.lua", t)
//}

//func TestMime(t *testing.T) {
//	doTest("mimetest.lua", t)
//}

//func TestMimeDot(t *testing.T) {
//	doTest("stufftest.lua", t)
//}

//func TestSmtp(t *testing.T) {
//	doTest("smtptest.lua", t)
//}

//func TestSmtpMessage(t *testing.T) {
//	doTest("testmesg.lua", t)
//}

func TestSocketGetAddrInfo(t *testing.T) {
	doTest("test_getaddrinfo.lua", t)
}

//func TestSocketError(t *testing.T) {
//	doTest("test_socket_error.lua", t)
//}

func TestUrl(t *testing.T) {
	doTest("urltest.lua", t)
}

// ----------------------------------------------------------------------------
func doTest(testScript string, t *testing.T) {
	assertions := assert.New(t)

	// Bring up a GopherLua VM
	luaState := lua.NewState()
	defer luaState.Close()

	// Register the Gluasocket module
	socket.Preload(luaState)

	// Change working directory to where scripts are, so that nested scripts are found
	_=os.Chdir("testdata/luasocket-test")

	// Run Lua test script
	fmt.Printf("Running test %s\n", testScript)
	assertions.NoError(luaState.DoFile(testScript))
}
