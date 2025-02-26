package workspace

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/types"
	"github.com/xmapst/AutoExecFlow/internal/utils"
)

// Delete
// @Summary		删除
// @Description	删除指定目录或文件
// @Tags		工作目录
// @Accept		application/json
// @Produce		application/json
// @Param		task path string true "任务名称"
// @Param		path query string false "路径"
// @Success		200 {object} types.SBase[any]
// @Failure		500 {object} types.SBase[any]
// @Router		/api/v1/task/{task}/workspace [delete]
func Delete(c *gin.Context) {
	task := c.Param("task")
	if task == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	prefix := filepath.Join(config.App.WorkSpace(), task)
	if !utils.FileOrPathExist(prefix) {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	path := filepath.Join(prefix, utils.PathEscape(c.Query("path")))
	if err := os.RemoveAll(path); err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeFailed).WithError(err))
		return
	}
	base.Send(c, base.WithCode[any](types.CodeSuccess))
}
