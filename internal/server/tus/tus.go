package tus

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/tus"
	"github.com/xmapst/AutoExecFlow/pkg/tus/locker"
	memorylocker "github.com/xmapst/AutoExecFlow/pkg/tus/locker/memory"
	redislocker "github.com/xmapst/AutoExecFlow/pkg/tus/locker/redis"
	filestore "github.com/xmapst/AutoExecFlow/pkg/tus/storage/file"
	"github.com/xmapst/AutoExecFlow/pkg/tus/types"
)

var TunServer *tus.STusx

func Init(uploadDir, relativePath, redisUrl string) error {
	var _locker locker.ILocker
	if redisUrl != "" {
		var err error
		_locker, err = redislocker.New(redisUrl)
		if err != nil {
			logx.Errorln(err)
		}
	} else {
		_locker = memorylocker.New()
	}
	store, err := filestore.New(filepath.Join(os.TempDir(), ".tusd"), _locker)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	TunServer, err = tus.New(&tus.SConfig{
		BasePath: relativePath,
		Store:    store,
		Logger:   logx.GetSubLogger(),
		PreUploadCreateCallback: func(hook types.HookEvent) (types.HTTPResponse, types.FileInfoChanges, error) {
			id := ksuid.New().String()
			taskID, ok := hook.Upload.MetaData["taskid"]
			if !ok {
				return types.HTTPResponse{
					StatusCode: http.StatusBadRequest,
					Body:       "taskid is required",
				}, types.FileInfoChanges{}, errors.New("taskid is required")
			}
			return types.HTTPResponse{}, types.FileInfoChanges{
				ID: filepath.Join(taskID, id),
			}, nil
		},
		PreFinishResponseCallback: func(hook types.HookEvent) (types.HTTPResponse, error) {
			if hook.Upload.IsFinal {
				filename := hook.Upload.MetaData["filename"]
				if filename == "" {
					filename = filepath.Base(hook.Upload.ID)
				}

				src := filepath.Join(os.TempDir(), ".tusd", hook.Upload.ID)
				dst := filepath.Join(uploadDir, filepath.Dir(hook.Upload.ID), filename)
				if err = utils.CopyFile(src, dst); err != nil {
					return types.HTTPResponse{
						StatusCode: http.StatusInternalServerError,
						Body:       "failed to copy file",
					}, err
				}
			}
			return types.HTTPResponse{
				Headers: map[string]string{
					"ID":   filepath.Base(hook.Upload.ID),
					"Path": filepath.Dir(hook.Upload.ID),
				},
			}, nil
		},
	})
	if err != nil {
		logx.Errorln(err)
		return err
	}
	return nil
}

func Shutdown(ctx context.Context) {
	if TunServer != nil {
		_ = TunServer.Close(ctx)
	}
}
