package worker

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/expr-lang/expr"
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

		// 日志写入
		s.storage.Log().Write(common.ConsoleStart)
		defer s.storage.Log().Write(common.ConsoleDone)

		// 评估规则, 使用expr
		var action common.Action
		action, err = s.evaluateExprRule()
		if err != nil {
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			// 写入步骤日志
			s.storage.Log().Write(err.Error())
			res.State = models.Pointer(models.StateFailed)
			res.Code = models.Pointer(common.CodeSystemErr)
			res.Message = err.Error()
			return err
		}
		switch action {
		case common.ActionBlock:
			res.State = models.Pointer(models.StateBlocked)
			res.Code = models.Pointer(common.CodeBlocked)
			res.Message = "blocked due to rule"
			return errors.New(res.Message)
		case common.ActionSkip:
			res.State = models.Pointer(models.StateSkipped)
			res.Code = models.Pointer(common.CodeSkipped)
			res.Message = "skipped due to rule"
			return
		default:
			// 继续执行
		}

		s.before(ctx, taskName, stepName)
		defer func() {
			s.after(ctx, taskName, stepName)
		}()

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

func (s *sStep) evaluateExprRule() (common.Action, error) {
	// 查询规则
	rule, err := s.storage.Rule()
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		return common.ActionBlock, err
	}
	action, err := s.storage.Action()
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		return common.ActionBlock, err
	}
	if rule == "" || action == "" {
		logx.Infoln(s.storage.TaskName(), s.storage.Name(), "no rule or no action")
		return common.ActionAllow, nil
	}
	program, err := expr.Compile(
		rule,
		// 预期返回值类型
		expr.AsBool(),
		// TODO: 内置函数或工具链
		//expr.Env(map[string]any{
		//	"storage": s.storage,
		//}),
		//expr.Function("test", func(params ...any) (any, error) {
		//	return "test", nil
		//}),
	)
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		return common.ActionBlock, err
	}
	result, err := expr.Run(program, nil)
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		return common.ActionBlock, err
	}
	matched, ok := result.(bool)
	if !ok {
		return common.ActionBlock, fmt.Errorf("rule result is not a boolean")
	}
	// 如果不匹配, 继续执行
	if !matched {
		return common.ActionAllow, nil
	}
	return common.ActionConvert(action), nil
}

func (s *sStep) before(ctx context.Context, taskName, stepName string) {
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "started")
	return
}

func (s *sStep) after(ctx context.Context, taskName, stepName string) {
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "end")
	return
}
