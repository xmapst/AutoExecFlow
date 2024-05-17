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

type step struct {
	storage   backend.IStep
	wg        *sync.WaitGroup
	workspace string
	scriptDir string
	logChan   chan string
}

func (s *step) vertexFunc() dag.VertexFunc {
	// build step
	return func(ctx context.Context, taskName, stepName string) error {
		s.storage = storage.Task(taskName).Step(stepName)
		defer func() {
			go func() {
				s.wg.Wait()
				close(s.logChan)
			}()
			_err := recover()
			if _err == nil {
				return
			}
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), _err)
			if err := s.storage.Update(&models.StepUpdate{
				State:    models.Pointer(models.Failed),
				OldState: models.Pointer(models.Running),
				Code:     models.Pointer(exec.SystemErr),
				Message:  fmt.Sprint(_err),
				ETime:    models.Pointer(time.Now()),
			}); err != nil {
				logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			}
		}()
		var err error

		// Asynchronous processing of log output
		go s.writeLog()

		s.before(ctx, taskName, stepName)
		defer func() {
			s.after(ctx, taskName, stepName)
		}()

		if err = s.storage.Update(&models.StepUpdate{
			State:    models.Pointer(models.Running),
			OldState: models.Pointer(models.Pending),
			Message:  "step is running",
			STime:    models.Pointer(time.Now()),
		}); err != nil {
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			return err
		}
		_type, err := s.storage.Type()
		if err != nil {
			logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
			_ = s.storage.Update(&models.StepUpdate{
				State:    models.Pointer(models.Failed),
				OldState: models.Pointer(models.Running),
				Code:     models.Pointer(exec.SystemErr),
				Message:  err.Error(),
				ETime:    models.Pointer(time.Now()),
			})
			return err
		}
		switch _type {
		// TODO: other type
		default:
			if err = s.execStep(ctx, taskName, stepName); err != nil {
				logx.Errorln(s.storage.TaskName(), s.storage.Name(), err)
				return err
			}
		}
		return nil
	}
}

func (s *step) before(ctx context.Context, taskName, stepName string) {
	s.wg.Add(1)
	defer s.wg.Done()
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "started")
	return
}

func (s *step) after(ctx context.Context, taskName, stepName string) {
	s.wg.Add(1)
	defer s.wg.Done()
	logx.Infoln(s.storage.TaskName(), s.storage.Name(), s.workspace, s.scriptDir, "end")
	return
}

const ConsoleStart = "OSREAPI::CONSOLE::START"
const ConsoleDone = "OSREAPI::CONSOLE::DONE"

func (s *step) writeLog() {
	var num int64
	// start
	if err := s.storage.Log().Insert(&models.Log{
		Timestamp: time.Now().UnixNano(),
		Line:      models.Pointer(num),
		Content:   ConsoleStart,
	}); err != nil {
		logx.Warnln(s.storage.TaskName(), s.storage.Name(), err)
	}
	defer func() {
		// end
		num += 1
		if err := s.storage.Log().Insert(&models.Log{
			Timestamp: time.Now().UnixNano(),
			Line:      models.Pointer(num),
			Content:   ConsoleDone,
		}); err != nil {
			logx.Warnln(s.storage.TaskName(), s.storage.Name(), err)
		}
	}()
	// content
	for log := range s.logChan {
		logx.Debugln(s.storage.TaskName(), s.storage.Name(), log)
		// TODO: 从输出中获取内容设置到环境变量中心

		num += 1
		log = strings.ReplaceAll(log, ConsoleStart, "")
		log = strings.ReplaceAll(log, ConsoleDone, "")
		if err := s.storage.Log().Insert(&models.Log{
			Timestamp: time.Now().UnixNano(),
			Line:      models.Pointer(num),
			Content:   log,
		}); err != nil {
			logx.Warnln(s.storage.TaskName(), s.storage.Name(), err)
		}
	}
}
