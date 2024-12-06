package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/ioutil"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func Test_Api(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		inspect.Preload,
		time.Preload,
		ioutil.Preload,
		Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
