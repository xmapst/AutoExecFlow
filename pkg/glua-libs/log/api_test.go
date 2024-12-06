package log

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/filepath"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/ioutil"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/strings"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
)

func Test_Api(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		ioutil.Preload,
		Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}

func Test_LogLevelApi(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		ioutil.Preload,
		filepath.Preload,
		strings.Preload,
		Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_loglevel.lua"))
}
