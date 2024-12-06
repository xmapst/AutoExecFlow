package cert_util

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/tests"
)

func runHttps(addr string, handler http.Handler) *http.Server {
	server := &http.Server{Addr: addr, Handler: handler}
	go func() {
		_ = server.ListenAndServeTLS("./test/cert.pem", "./test/key.pem")
	}()
	return server
}

func httpRouterGet(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, "OK")
}

func Test_Api(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/get", httpRouterGet)
	server := runHttps(":1443", mux)
	t.Cleanup(func() {
		_ = server.Shutdown(context.Background())
	})
	time.Sleep(time.Second)

	assert.NotZero(t, tests.RunLuaTestFile(t, Preload, "./test/test_api.lua"))
}
