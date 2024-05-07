package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/xmapst/osreapi/internal/storage"
	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/dag"
	"github.com/xmapst/osreapi/pkg/exec"
	"github.com/xmapst/osreapi/pkg/logx"
)

type Step struct {
	backend.IStep
	wg        *sync.WaitGroup
	scriptDir string
	workspace string
	logChan   chan string

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
			Code:     models.Pointer(int64(0)),
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
		s.logChan = make(chan string, 65535)
		defer func() {
			go func() {
				s.wg.Wait()
				close(s.logChan)
			}()
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
				ETime:    models.Pointer(time.Now()),
			}); err != nil {
				logx.Errorln(s.TaskName, s.Name, s.workspace, s.scriptDir, err)
			}
		}()
		var err error
		// Asynchronous processing of log output
		go s.writeLog()

		s.before()
		defer func() {
			s.after()
		}()

		if err = s.Update(&models.StepUpdate{
			State:    models.Pointer(models.Running),
			OldState: models.Pointer(models.Pending),
			Message:  "step is running",
			STime:    models.Pointer(time.Now()),
		}); err != nil {
			logx.Errorln(err)
			return err
		}

		switch s.Type {
		// TODO: other type
		default:
			if err = s.execStep(ctx, globalEnv); err != nil {
				logx.Errorln(err)
				return err
			}
		}
		return nil
	}, nil
}

func (s *Step) before() {
	s.wg.Add(1)
	defer s.wg.Done()
	logx.Infoln(s.TaskName, s.Name, s.workspace, s.scriptDir, "started")
	return
}

func (s *Step) after() {
	s.wg.Add(1)
	defer s.wg.Done()
	logx.Infoln(s.TaskName, s.Name, s.workspace, s.scriptDir, "end")
	return
}

const ConsoleStart = "OSREAPI::CONSOLE::START"
const ConsoleDone = "OSREAPI::CONSOLE::DONE"

func (s *Step) writeLog() {
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
	for log := range s.logChan {
		logx.Debugln(log)
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
