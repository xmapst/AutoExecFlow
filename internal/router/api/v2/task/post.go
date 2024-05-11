package taskv2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var req = new(types.Task)
	if err := base.DecodeJson(r.Body, req); err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(err))
		return
	}

	if err := req.Save(); err != nil {
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(err))
		return
	}

	r.Header.Set(types.XTaskName, req.Name)
	w.Header().Set(types.XTaskName, req.Name)

	var scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	path := strings.Replace(strings.TrimSuffix(r.URL.Path, "/"), "v2", "v1", 1)
	base.SendJson(w, base.New().WithData(&types.TaskCreateRes{
		URL:   fmt.Sprintf("%s://%s%s/%s", scheme, r.Host, path, req.Name),
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	}))
}
