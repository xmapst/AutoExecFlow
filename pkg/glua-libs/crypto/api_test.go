package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
)

func Test_Api(t *testing.T) {
	assert.NotZero(t, tests.RunLuaTestFile(t, Preload, "./test/test_api.lua"))
}
