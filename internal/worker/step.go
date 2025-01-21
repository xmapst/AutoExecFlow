package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	"github.com/xmapst/AutoExecFlow/internal/worker/runner"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
)

type sStep struct {
	storage   storage.IStep
	workspace string
	scriptDir string
}

func newStep(storage storage.IStep, workspace, scriptDir string) dag.VertexFunc {
	s := &sStep{
		storage:   storage,
		workspace: workspace,
		scriptDir: scriptDir,
	}
	return s.vertexFunc()
}

func (s *sStep) vertexFunc() dag.VertexFunc {
	// build step
	return func(ctx context.Context, taskName, stepName string) (err error) {
		if err = s.storage.Update(&models.SStepUpdate{
			State:    models.Pointer(models.StateRunning),
			OldState: models.Pointer(models.StatePending),
			Message:  "step is running",
			STime:    models.Pointer(time.Now()),
		}); err != nil {
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			return err
		}

		// proc step
		var res = new(models.SStepUpdate)
		defer func() {
			if _err := recover(); _err != nil {
				stack := string(debug.Stack())
				logx.Errorln(stack, _err)
				res.Code = models.Pointer(common.CodeSystemErr)
				res.Message = fmt.Sprint(_err)
			}
			res.ETime = models.Pointer(time.Now())
			res.OldState = models.Pointer(models.StateRunning)
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
			res.State = models.Pointer(models.StateFailed)
			res.Message = err.Error()
			res.Code = models.Pointer(common.CodeSystemErr)
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
		res.State = models.Pointer(models.StateStopped)
		if err != nil {
			res.State = models.Pointer(models.StateFailed)
			res.Message = err.Error()
			return err
		}
		if code != 0 {
			res.State = models.Pointer(models.StateFailed)
			res.Message = fmt.Sprintf("execution failed with code: %d", code)
			return errors.New(res.Message)
		}
		return nil
	}
}

func (s *sStep) before(ctx context.Context, taskName, stepName string) {
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "started")
	return
}

func (s *sStep) after(ctx context.Context, taskName, stepName string) {
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "end")
	return
}
