package glualibs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
)

func Test_Preload(t *testing.T) {
	assert.NotZero(t, tests.RunLuaTestFile(t, Preload, "./preload.lua"))
}
