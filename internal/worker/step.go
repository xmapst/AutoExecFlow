package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/xmapst/AutoExecFlow/internal/runner"
	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type step struct {
	storage   backend.IStep
	workspace string
	scriptDir string
}

func (s *step) vertexFunc() dag.VertexFunc {
	// build step
	return func(ctx context.Context, taskName, stepName string) (err error) {
		if err = s.storage.Update(&models.StepUpdate{
			State:    models.Pointer(models.Running),
			OldState: models.Pointer(models.Pending),
			Message:  "step is running",
			STime:    models.Pointer(time.Now()),
		}); err != nil {
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			return err
		}

		// proc step
		var res = new(models.StepUpdate)
		defer func() {
			if _err := recover(); _err != nil {
				logx.Errorln(_err)
				res.Code = models.Pointer(common.SystemErr)
				res.Message = fmt.Sprint(_err)
			}
			res.ETime = models.Pointer(time.Now())
			res.OldState = models.Pointer(models.Running)
			if _err := s.storage.Update(res); _err != nil {
				logx.Errorln(_err)
			}
		}()

		s.before(ctx, taskName, stepName)
		defer func() {
			s.after(ctx, taskName, stepName)
		}()

		// 日志写入
		s.storage.Log().Write(common.ConsoleStart)
		defer s.storage.Log().Write(common.ConsoleDone)

		runnerItr, err := runner.New(
			s.storage,
			s.workspace,
			s.scriptDir,
		)
		if err != nil {
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			res.State = models.Pointer(models.Failed)
			res.Message = err.Error()
			res.Code = models.Pointer(common.SystemErr)
			return err
		}

		defer func() {
			if cErr := runnerItr.Clear(); cErr != nil {
				logx.Warnln(cErr)
			}
		}()

		res.Message = "execution succeed"
		code, err := runnerItr.Run(ctx)
		res.Code = models.Pointer(code)
		res.State = models.Pointer(models.Stop)
		if err != nil {
			res.State = models.Pointer(models.Failed)
			res.Message = err.Error()
			return err
		}
		if code != 0 {
			res.State = models.Pointer(models.Failed)
			res.Message = fmt.Sprintf("execution failed with code: %d", code)
			return errors.New(res.Message)
		}
		return nil
	}
}

func (s *step) before(ctx context.Context, taskName, stepName string) {
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "started")
	return
}

func (s *step) after(ctx context.Context, taskName, stepName string) {
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "end")
	return
}
