package pb

import (
	"testing"

	lua "github.com/yuin/gopher-lua"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func Test_AllParams(t *testing.T) {
	state := lua.NewState()
	Preload(state)
	time.Preload(state)
	source := `
                local pb = require('pb')
                local time = require('time')

                local count = 2
                local bar = pb.new(count)
                local template = string.format('%s {{ counters . }} {{percent . }} {{ etime . }}', '[custom template]')

                err = bar:configure({writer='stdout', refresh_rate=3001, template=template})
                if err then error(err) end
                bar:start()

                for i=1, count, 1 do
                  time.sleep(1)
                  bar:increment()
                end
                bar:finish()
        `
	if err := state.DoString(source); err != nil {
		t.Fatal(err.Error())
	}
	// Output:
	// [custom template] 2 / 2 100.00% 2s

}
