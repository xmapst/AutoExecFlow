package pprof_test

import (
	"log"

	lua "github.com/yuin/gopher-lua"

	luahttp "github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
	luapprof "github.com/xmapst/AutoExecFlow/pkg/glua-libs/pprof"
	luatime "github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

// pprof:register(), pprof_ud:enable(), pprof_ud:disable()
func Example_package() {
	state := lua.NewState()
	luapprof.Preload(state)
	luahttp.Preload(state)
	luatime.Preload(state)
	source := `
local pprof = require("pprof")
local http = require("http")
local time = require("time")

local client = http.client()
local pp = pprof.register(":1234")

pp:enable()
time.sleep(1)

local req, err = http.request("GET", "http://127.0.0.1:1234/debug/pprof/goroutine")
if err then error(err) end
local resp, err = client:do_request(req)
if err then error(err) end
if not(resp.code == 200) then error("resp code") end
print(resp.code)

pp:disable()
time.sleep(5)

local resp, err = client:do_request(req)
if not(err) then error("must be error") end
        `
	if err := state.DoString(source); err != nil {
		log.Fatal(err.Error())
	}
	// Output:
	// 200
}
