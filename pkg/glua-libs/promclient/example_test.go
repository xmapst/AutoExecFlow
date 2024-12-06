package promclient_test

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/promclient"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

// prometheus:start(string)
func Test_Start(t *testing.T) {
	state := lua.NewState()
	promclient.Preload(state)
	time.Preload(state)
	http.Preload(state)

	source := `
    local prometheus = require("prometheus")
	local time = require("time")
	local http = require("http_client")

	local pp = prometheus.register(":18080")
	pp:start()
    time.sleep(1)

	local client = http.client({timeout=5})

	local request = http.request("GET", "http://127.0.0.1:18080/")
	local result = client:do_request(request)
	print(result.code)

	local request = http.request("GET", "http://127.0.0.1:18080/metrics")
	local result = client:do_request(request)
	print(result.code)
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// 404
	// 200
}
