package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
)

type Step struct {
	backend.IStep
	scriptDir string
	workspace string

	TaskName string
	Name     string
	Type     string
	Content  string
	Timeout  time.Duration
}

func (s *Step) SaveEnv(env map[string]string) (err error) {
	if len(env) == 0 {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		storage.Task(s.TaskName).ClearAll()
	}()
	var envs []*models.Env
	for name, value := range env {
		envs = append(envs, &models.Env{
			Name:  name,
			Value: value,
		})
	}
	err = storage.Task(s.TaskName).Step(s.Name).Env().Create(envs)
	if err != nil {
		return err
	}
	return
}

func (s *Step) SaveDepends(depends []string) (err error) {
	if len(depends) == 0 {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		storage.Task(s.TaskName).ClearAll()
	}()
	err = storage.Task(s.TaskName).Step(s.Name).Depend().Create(depends)
	if err != nil {
		return err
	}
	return
}

func (s *Step) build(globalEnv backend.IEnv) (dag.VertexFunc, error) {
	// 设置缓存中初始状态
	if err := s.Create(&models.Step{
		Type:    s.Type,
		Content: s.Content,
		Timeout: s.Timeout,
		StepUpdate: models.StepUpdate{
			State:    models.Pointer(models.Pending),
			OldState: models.Pointer(models.Pending),
			Message:  "The current step only proceeds if the previous step succeeds.",
		},
	}); err != nil {
		logx.Errorln(s.TaskName, s.Name, s.workspace, s.scriptDir, err)
		return nil, err
	}
	// build step
	return func(ctx context.Context, taskName, stepName string) error {
		defer func() {
			_err := recover()
			if _err == nil {
				return
			}
			logx.Errorln(s.TaskName, s.Name, s.workspace, s.scriptDir, _err)
			if err := s.Update(&models.StepUpdate{
				State:    models.Pointer(models.Failed),
				OldState: models.Pointer(models.Running),
				Code:     models.Pointer(exec.SystemErr),
				Message:  fmt.Sprint(_err),
				ETime:    time.Now(),
			}); err != nil {
				logx.Errorln(s.TaskName, s.Name, s.workspace, s.scriptDir, err)
			}
		}()
		var err error
		// TODO: 执行前
		logx.Infoln(s.TaskName, s.Name, s.workspace, s.scriptDir, "started")
		defer func() {
			// TODO: 执行后
			logx.Infoln(s.TaskName, s.Name, s.workspace, s.scriptDir, "end")
		}()

		if err = s.execStep(ctx, globalEnv, s.workspace, s.scriptDir); err != nil {
			logx.Errorln(err)
			return err
		}
		return nil
	}, nil
}

const ConsoleStart = "OSREAPI::CONSOLE::START"
const ConsoleDone = "OSREAPI::CONSOLE::DONE"

func (s *Step) execStep(ctx context.Context, globalEnv backend.IEnv, workspace, scriptDir string) error {
	if err := s.Update(&models.StepUpdate{
		State:    models.Pointer(models.Running),
		OldState: models.Pointer(models.Pending),
		Message:  "step is running",
		STime:    time.Now(),
	}); err != nil {
		logx.Errorln(err)
		return err
	}

	var res = new(models.StepUpdate)
	defer func() {
		if _err := recover(); _err != nil {
			logx.Errorln(_err)
			res.Code = models.Pointer(exec.SystemErr)
			res.Message = fmt.Sprint(_err)
		}
		res.ETime = time.Now()
		res.OldState = models.Pointer(models.Running)
		if _err := s.Update(res); _err != nil {
			logx.Errorln(_err)
		}
	}()
	var logCh = make(chan string, 65535)
	// 动态获取环境变量
	var envs = make([]string, 0)
	taskEnv := globalEnv.List()
	for _, env := range taskEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	stepEnv := s.Env().List()
	for _, env := range stepEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	var cmd, err = exec.New(
		exec.WithLogger(logx.GetSubLoggerWithOption(zap.AddCallerSkip(-1))),
		exec.WithEnv(append(envs,
			fmt.Sprintf("TASK_NAME=%s", s.TaskName),
			fmt.Sprintf("TASK_STEP_NAME=%s", s.Name)),
		),
		exec.WithShell(s.Type),
		exec.WithScript(s.Content),
		exec.WithWorkspace(workspace),
		exec.WithScriptDir(scriptDir),
		exec.WithTimeout(s.Timeout),
		exec.WithConsoleCh(logCh),
	)
	if err != nil {
		logx.Errorln(s.TaskName, s.Name, workspace, scriptDir, err)
		res.Message = err.Error()
		res.Code = models.Pointer(exec.SystemErr)
		return err
	}
	go s.writeLog(logCh)
	res.Message = "execution succeed"
	code, err := cmd.Run(ctx)
	res.Code = models.Pointer(code)
	res.State = models.Pointer(models.Stop)
	if err != nil {
		fmt.Println(err.Error())
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

func (s *Step) writeLog(logCh chan string) {
	var num int64
	// start
	if err := s.Log().Create(&models.Log{
		Timestamp: time.Now().UnixNano(),
		Line:      models.Pointer(num),
		Content:   ConsoleStart,
	}); err != nil {
		logx.Warnln(err)
	}
	defer func() {
		// end
		num += 1
		if err := s.Log().Create(&models.Log{
			Timestamp: time.Now().UnixNano(),
			Line:      models.Pointer(num),
			Content:   ConsoleDone,
		}); err != nil {
			logx.Warnln(err)
		}
	}()
	// content
	for log := range logCh {
		// TODO: 从输出中获取内容设置到环境变量中心

		num += 1
		log = strings.ReplaceAll(log, ConsoleStart, "")
		log = strings.ReplaceAll(log, ConsoleDone, "")
		if err := s.Log().Create(&models.Log{
			Timestamp: time.Now().UnixNano(),
			Line:      models.Pointer(num),
			Content:   log,
		}); err != nil {
			logx.Warnln(err)
		}
	}
}
