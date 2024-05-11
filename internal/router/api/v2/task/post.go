package taskv2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"

	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(w http.ResponseWriter, r *http.Request) {
	var req = new(types.Task)
	if err := render.DecodeJSON(r.Body, req); err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(err))
		return
	}

	if err := req.Save(); err != nil {
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(err))
		return
	}

	r.Header.Set(types.XTaskName, req.Name)
	w.Header().Set(types.XTaskName, req.Name)

	var scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	path := strings.Replace(strings.TrimSuffix(r.URL.Path, "/"), "v2", "v1", 1)
	render.JSON(w, r, types.New().WithData(&types.TaskCreateRes{
		URL:   fmt.Sprintf("%s://%s%s/%s", scheme, r.Host, path, req.Name),
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	}))
}
