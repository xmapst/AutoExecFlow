package workspace

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(w http.ResponseWriter, r *http.Request) {
	task := chi.URLParam(r, "task")
	if task == "" {
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	prefix := filepath.Join(config.App.WorkSpace, task)
	if !utils.FileOrPathExist(prefix) {
		base.SendJson(w, base.New().WithCode(base.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	path := filepath.Join(prefix, utils.PathEscape(r.URL.Query().Get("path")))
	if err := base.SaveFiles(r, path); err != nil {
		logx.Errorln(err)
		base.SendJson(w, base.New().WithCode(base.CodeFailed).WithError(err))
		return
	}
	base.SendJson(w, base.New())
}
