package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/goos"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/strings"
)

func TestSuite(t *testing.T) {
	preload := strings.Preload
	assert.NotZero(t, RunLuaTestFile(t, preload, "testdata/test_suite.lua"))
}

func TestApi(t *testing.T) {
	preload := goos.Preload
	assert.NotZero(t, RunLuaTestFile(t, preload, "testdata/test_api.lua"))
}

func TestAssertions(t *testing.T) {
	t.Run("passing", func(t *testing.T) {
		preload := inspect.Preload
		assert.NotZero(t, RunLuaTestFile(t, preload, "testdata/test_assertions_passing.lua"))
	})
	t.Run("failing", func(t *testing.T) {
		if _, ok := os.LookupEnv("TEST_ASSERTIONS_FAILING"); !ok {
			t.Skip("Skipping unless TEST_ASSERTIONS_FAILING is set")
		}
		preload := inspect.Preload
		assert.NotZero(t, RunLuaTestFile(t, preload, "testdata/test_assertions_failing.lua"))
	})
}
