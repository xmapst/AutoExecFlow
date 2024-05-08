package workspace

import (
	"errors"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/xmapst/osreapi/internal/router/base"
	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Post(c *gin.Context) {
	render := base.Gin{Context: c}
	task := c.Param("task")
	if task == "" {
		render.SetError(base.CodeNoData, errors.New("task does not exist"))
		return
	}
	prefix := filepath.Join(config.App.WorkSpace, task)
	if !utils.FileOrPathExist(prefix) {
		render.SetError(base.CodeNoData, errors.New("task does not exist"))
		return
	}
	path := filepath.Join(prefix, utils.PathEscape(c.Query("path")))
	if err := render.SaveFiles(path); err != nil {
		logx.Errorln(err)
		render.SetError(base.CodeFailed, err)
		return
	}
	render.SetRes(nil)
}
