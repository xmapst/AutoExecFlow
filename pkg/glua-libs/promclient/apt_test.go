package promclient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/promclient"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/strings"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func Test_Api(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		promclient.Preload,
		http.Preload,
		strings.Preload,
		time.Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
