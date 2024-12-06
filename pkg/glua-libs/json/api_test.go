package json

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/strings"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
)

func Test_Api(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		Preload,
		inspect.Preload,
		strings.Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
