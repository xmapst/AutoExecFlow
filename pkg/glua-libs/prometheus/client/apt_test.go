package prometheus_client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/http"
	prometheus "github.com/xmapst/AutoExecFlow/pkg/glua-libs/prometheus/client"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/strings"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/time"
)

func TestApi(t *testing.T) {
	preload := tests.SeveralPreloadFuncs(
		prometheus.Preload,
		http.Preload,
		strings.Preload,
		time.Preload,
	)
	assert.NotZero(t, tests.RunLuaTestFile(t, preload, "./test/test_api.lua"))
}
