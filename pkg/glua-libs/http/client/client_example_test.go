package http_test

import (
	"log"
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
)

// http_client_ud:do_request(request)
func Test_DoRequest(t *testing.T) {
	state := lua.NewState()
	http.Preload(state)
	source := `
    local http = require("http")
    local client = http.client()
    local request = http.request("GET", "https://google.com")
    local result, err = client:do_request(request)
    if err then error(err) end
    print(result.code)
`
	if err := state.DoString(source); err != nil {
		log.Fatal(err.Error())
	}
	// Output:
	// 200
}

// http:query_escape(string)
func Test_QueryEscape(t *testing.T) {
	state := lua.NewState()
	http.Preload(state)
	source := `
    local http = require("http")
    local result = http.query_escape("<> 123")
    print(result)
`
	if err := state.DoString(source); err != nil {
		log.Fatal(err.Error())
	}
	// Output:
	// %3C%3E+123
}

// http:query_escape(string)
func Test_QueryUnescape(t *testing.T) {
	state := lua.NewState()
	http.Preload(state)
	source := `
    local http = require("http")
    local result, err = http.query_unescape("%3C%3E+123")
    if err then error(err) end
    print(result)
`
	if err := state.DoString(source); err != nil {
		log.Fatal(err.Error())
	}
	// Output:
	// <> 123
}
