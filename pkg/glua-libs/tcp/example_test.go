package tcp

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// tcp.open(), tcp_client_ud:write(), tcp_client_ud:read()
func Test_full(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	go func() {
		_, err := runPingPongServer(":12346")
		if err != nil {
			t.Error(err)
			return
		}
	}()
	time.Sleep(time.Second)
	source := `
        local tcp = require("tcp")

        local conn, err = tcp.open(":12346")
        if err then error(err) end

        -- send ping, read "pong\n"
        local err = conn:write("ping")
        if err then error(err) end
        local result, err = conn:read()
        if err then error(err) end
        print(result)

        -- send ping, read by byte
        local err = conn:write("ping")
        if err then error(err) end
        for i = 1, 5 do
            local result, err = conn:read(1)
            if err then error(err) end
            print(result)
        end

        conn:close()
`
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// pong
	//
	// p
	// o
	// n
	// g
	//
}
