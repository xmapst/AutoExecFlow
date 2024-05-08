package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
)

func (s *step) execStep(ctx context.Context, taskName, stepName string) error {
	var res = new(models.StepUpdate)
	defer func() {
		if _err := recover(); _err != nil {
			logx.Errorln(_err)
			res.Code = models.Pointer(exec.SystemErr)
			res.Message = fmt.Sprint(_err)
		}
		res.ETime = models.Pointer(time.Now())
		res.OldState = models.Pointer(models.Running)
		if _err := s.storage.Update(res); _err != nil {
			logx.Errorln(_err)
		}
	}()

	// 动态获取环境变量
	var envs = make([]string, 0)
	taskEnv := storage.Task(taskName).Env().List()
	for _, env := range taskEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	stepEnv := s.storage.Env().List()
	for _, env := range stepEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	var logChan = make(chan string)
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()
		for log := range logChan {
			s.logChan <- log
		}
	}()
	_type, err := s.storage.Type()
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		res.Code = models.Pointer(exec.SystemErr)
		return err
	}
	content, err := s.storage.Content()
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		res.Code = models.Pointer(exec.SystemErr)
		return err
	}
	timeout, err := s.storage.Timeout()
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		res.Code = models.Pointer(exec.SystemErr)
		return err
	}
	cmd, err := exec.New(
		exec.WithEnv(
			append(
				envs,
				fmt.Sprintf("TASK_NAME=%s", s.storage.TaskName()),
				fmt.Sprintf("TASK_STEP_NAME=%s", s.storage.Name()),
				fmt.Sprintf("TASK_WORKSPACE=%s", s.workspace),
				fmt.Sprintf("TASK_SCRIPT_DIR=%s", s.scriptDir),
			),
		),
		exec.WithShell(_type),
		exec.WithScript(content),
		exec.WithWorkspace(s.workspace),
		exec.WithScriptDir(s.scriptDir),
		exec.WithTimeout(timeout),
		exec.WithConsoleCh(logChan),
	)
	if err != nil {
		logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
		res.State = models.Pointer(models.Failed)
		res.Message = err.Error()
		res.Code = models.Pointer(exec.SystemErr)
		return err
	}

	defer func() {
		// clear tmp script
		if cErr := cmd.Clear(); cErr != nil {
			logx.Warnln(cErr)
		}
	}()

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
