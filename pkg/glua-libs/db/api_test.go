//go:build !windows && sqlite
// +build !windows,sqlite

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/inspect"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func TestApi(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		time.Preload,
		inspect.Preload,
		Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
