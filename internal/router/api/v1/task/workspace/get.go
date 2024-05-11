package workspace

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/xmapst/osreapi/internal/router/types"
	"github.com/xmapst/osreapi/internal/server/config"
	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/logx"
)

func Get(w http.ResponseWriter, r *http.Request) {
	taskName := chi.URLParam(r, "task")
	if taskName == "" {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	prefix := filepath.Join(config.App.WorkSpace, taskName)
	if !utils.FileOrPathExist(prefix) {
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(errors.New("task does not exist")))
		return
	}
	path := filepath.Join(prefix, utils.PathEscape(r.URL.Query().Get("path")))
	file, err := os.Open(path)
	if err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
		return
	}
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		finalPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			logx.Errorln(err)
			render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
			return
		}
		fileInfo, err = os.Lstat(finalPath)
		if err != nil {
			logx.Errorln(err)
			render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
			return
		}
	}
	if !fileInfo.IsDir() {
		ctype := mime.TypeByExtension(fileInfo.Name())
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ctype)
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name()))
		_, _ = io.Copy(w, file)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		logx.Errorln(err)
		render.JSON(w, r, types.New().WithCode(types.CodeNoData).WithError(err))
		return
	}
	var scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}

	var infos = new(types.FileListRes)
	var uriPrefix = fmt.Sprintf("%s://%s%s", scheme, r.Host, strings.TrimSuffix(r.URL.Path, "/"))
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
			URL:     fmt.Sprintf("%s?path=%s", uriPrefix, _path),
			Name:    info.Name(),
			Path:    _path,
			Size:    size,
			Mode:    info.Mode().String(),
			ModTime: info.ModTime().Unix(),
			IsDir:   isDir,
		})
	}
	infos.Total = len(infos.Files)
	render.JSON(w, r, types.New().WithData(infos))
}
