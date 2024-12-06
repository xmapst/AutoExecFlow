package pprof_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	luahttp "github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
	luapprof "github.com/xmapst/AutoExecFlow/pkg/glua-libs/pprof"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
	luatime "github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func Test_Api(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		luapprof.Preload,
		luahttp.Preload,
		luatime.Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
