package pb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func Test_Api(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		time.Preload,
		Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
