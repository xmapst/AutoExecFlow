package workspace

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	task := chi.URLParam(r, "task")
	if task == "" {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	prefix := filepath.Join(config.App.WorkSpace, task)
	if !utils.FileOrPathExist(prefix) {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	path := filepath.Join(prefix, utils.PathEscape(r.URL.Query().Get("path")))
	if err := os.RemoveAll(path); err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeFailed).WithError(err))
		return
	}
	render.JSON(w, r, types.New())
}
