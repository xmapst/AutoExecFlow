package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
)

func (s *Step) execStep(ctx context.Context, globalEnv backend.IEnv) error {
	var res = new(models.StepUpdate)
	defer func() {
		if _err := recover(); _err != nil {
			logx.Errorln(_err)
			res.Code = models.Pointer(exec.SystemErr)
			res.Message = fmt.Sprint(_err)
		}
		res.ETime = models.Pointer(time.Now())
		res.OldState = models.Pointer(models.Running)
		if _err := s.Update(res); _err != nil {
			logx.Errorln(_err)
		}
	}()

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
	var logChan = make(chan string)
	go func() {
		s.wg.Add(1)
		defer s.wg.Done()
		for log := range logChan {
			s.logChan <- log
		}
	}()

	var cmd, err = exec.New(
		exec.WithEnv(
			append(
				envs,
				fmt.Sprintf("TASK_NAME=%s", s.TaskName),
				fmt.Sprintf("TASK_STEP_NAME=%s", s.Name),
				fmt.Sprintf("TASK_WORKSPACE=%s", s.workspace),
				fmt.Sprintf("TASK_SCRIPT_DIR=%s", s.scriptDir),
			),
		),
		exec.WithShell(s.Type),
		exec.WithScript(s.Content),
		exec.WithWorkspace(s.workspace),
		exec.WithScriptDir(s.scriptDir),
		exec.WithTimeout(s.Timeout),
		exec.WithConsoleCh(logChan),
	)
	if err != nil {
		logx.Errorln(s.TaskName, s.Name, s.workspace, s.scriptDir, err)
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
