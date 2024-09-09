package workspace

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

// Get
// @Summary		Get
// @description	Get file/folder contents
// @Tags		Workspace
// @Accept		application/json
// @Accept		application/yaml
// @Produce		application/json
// @Produce		application/yaml
// @Param		task path string true "task name"
// @Param		path query string false "dir path"
// @Success		200 {object} types.Base[types.FileListRes]
// @Failure		500 {object} types.Base[any]
// @Router		/api/v1/task/{task}/workspace [get]
func Get(c *gin.Context) {
	taskName := c.Param("task")
	if taskName == "" {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	prefix := filepath.Join(config.App.WorkSpace(), taskName)
	if !utils.FileOrPathExist(prefix) {
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	path := filepath.Join(prefix, utils.PathEscape(c.Query("path")))
	file, err := os.Open(path)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		finalPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			logx.Errorln(err)
			base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
			return
		}
		fileInfo, err = os.Lstat(finalPath)
		if err != nil {
			logx.Errorln(err)
			base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
			return
		}
	}
	if !fileInfo.IsDir() {
		ctype := mime.TypeByExtension(fileInfo.Name())
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		c.Header("Content-Type", ctype)
		c.Header("Transfer-Encoding", "chunked")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name()))
		_, _ = io.Copy(c.Writer, file)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		logx.Errorln(err)
		base.Send(c, base.WithCode[any](types.CodeNoData).WithError(err))
		return
	}

	var infos = new(types.FileListRes)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		_path := filepath.Join(path, info.Name())
		isDir := info.IsDir()
		size := info.Size()
		if info.Mode()&os.ModeSymlink != 0 {
			s, err := os.Stat(_path)
			if err == nil {
				isDir = s.IsDir()
				size = s.Size()
			}
		}
		_path = strings.TrimPrefix(filepath.ToSlash(_path), filepath.ToSlash(prefix))
		infos.Files = append(infos.Files, &types.FileRes{
			Name:    info.Name(),
			Path:    _path,
			Size:    size,
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().UnixNano(),
			IsDir:   isDir,
		})
	}
	infos.Total = len(infos.Files)
	base.Send(c, base.WithData(infos))
}
