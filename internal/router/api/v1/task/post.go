package task

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
	for k, v := range r.URL.Query() {
		switch k {
		case "name":
			req.Name = v[0]
		case "timeout":
			req.Timeout = v[0]
		case "async":
			req.Async = v[0] == "true"
		case "env_vars":
			req.EnvVars = v
		case "env":
			if req.Env == nil {
				req.Env = make(map[string]string)
			}
			for _, str := range v {
				before, after, found := strings.Cut(str, ":")
				if !found {
					continue
				}
				req.Env[before] = after
			}
		}
	}

	if err := base.DecodeJson(r.Body, &req.Step); err != nil {
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
	base.SendJson(w, base.New().WithData(&types.TaskCreateRes{
		URL:   fmt.Sprintf("%s://%s%s/%s", scheme, r.Host, strings.TrimSuffix(r.URL.Path, "/"), req.Name),
		ID:    req.Name,
		Name:  req.Name,
		Count: len(req.Step),
	}))
}
