package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/api/base"
	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/dag"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

type StepService struct {
	taskName string
	stepName string
}

func Step(taskName string, stepName string) *StepService {
	return &StepService{
		taskName: taskName,
		stepName: stepName,
	}
}

func (ss *StepService) Create(globalTimeout time.Duration, step *types.TaskStepReq) error {
	if globalTimeout <= 0 {
		return errors.New("global timeout must be greater than 0")
	}
	timeout, err := ss.review(step)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	if timeout <= 0 || timeout > globalTimeout {
		timeout = globalTimeout
	}
	var db = storage.Task(ss.taskName).Step(step.Name)
	if err = ss.saveStep(db, timeout, step); err != nil {
		return err
	}
	return nil
}

func (ss *StepService) review(step *types.TaskStepReq) (time.Duration, error) {
	step.Name = reg.ReplaceAllString(step.Name, "")
	if step.Name == "" {
		step.Name = ksuid.New().String()
	}
	if step.CommandType != "" && step.Type == "" {
		step.Type = step.CommandType
	}
	if step.Type == "" {
		return 0, errors.New("key: 'Step.Type' Error:Field validation for 'Type' failed on the 'required' tag")
	}

	if step.CommandContent != "" && step.Content == "" {
		step.Content = step.CommandContent
	}
	if step.Content == "" {
		return 0, errors.New("key: 'Step.Content' Error:Field validation for 'Content' failed on the 'required' tag")
	}
	if step.Env == nil {
		step.Env = make(map[string]string)
	}
	for kk, vv := range utils.SliceToStrMap(step.EnvVars) {
		if _, ok := step.Env[kk]; !ok {
			step.Env[kk] = vv
		}
	}

	timeout, _ := time.ParseDuration(step.Timeout)
	return timeout, nil
}

func (ss *StepService) saveStep(db backend.IStep, timeout time.Duration, step *types.TaskStepReq) (err error) {
	defer func() {
		if err != nil {
			db.ClearAll()
		}
	}()
	err = db.Insert(&models.Step{
		Type:    step.Type,
		Content: step.Content,
		Timeout: timeout,
		Disable: models.Pointer(step.Disable),
		StepUpdate: models.StepUpdate{
			Message:  "the step is waiting to be scheduled for execution",
			Code:     models.Pointer(int64(0)),
			State:    models.Pointer(models.Pending),
			OldState: models.Pointer(models.Pending),
		},
	})
	if err != nil {
		return err
	}
	// save step env
	for name, value := range step.Env {
		if err = db.Env().Insert(&models.Env{
			Name:  name,
			Value: value,
		}); err != nil {
			return err
		}
	}
	// save step depend
	err = db.Depend().Insert(step.Depends...)
	if err != nil {
		return err
	}
	return nil
}

func (ss *StepService) Detail() (types.Code, *types.TaskStepRes, error) {
	db := storage.Task(ss.taskName).Step(ss.stepName)
	step, err := db.Get()
	if err != nil {
		logx.Errorln(err)
		return types.CodeFailed, nil, err
	}
	data := &types.TaskStepRes{
		Name:    step.Name,
		State:   models.StateMap[*step.State],
		Code:    *step.Code,
		Message: step.Message,
		Timeout: step.Timeout.String(),
		Disable: *step.Disable,
		Env:     make(map[string]string),
		Type:    step.Type,
		Content: step.Content,
		Time: &types.TimeRes{
			Start: step.STimeStr(),
			End:   step.ETimeStr(),
		},
	}
	data.Depends = storage.Task(ss.taskName).Step(step.Name).Depend().List()
	envs := db.Env().List()
	for _, env := range envs {
		data.Env[env.Name] = env.Value
	}
	return types.Code(data.Code), data, nil
}

func (ss *StepService) Manager(action string, duration string) error {
	manager, err := dag.VertexManager(ss.taskName, ss.stepName)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	step, err := storage.Task(ss.taskName).Step(ss.stepName).Get()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	if *step.State != models.Running && *step.State != models.Pending && *step.State != models.Paused {
		return errors.New("step is no running")
	}
	switch action {
	case "kill":
		err = manager.Kill()
		if err == nil {
			return storage.Task(ss.taskName).Step(ss.stepName).Update(&models.StepUpdate{
				State:    models.Pointer(models.Failed),
				OldState: step.State,
				Message:  "has been killed",
			})
		}
	case "pause":
		if *step.State == models.Running {
			return dag.ErrRunning
		}
		if manager.State() != dag.Paused {
			_ = manager.Pause(duration)
			return storage.Task(ss.taskName).Step(ss.stepName).Update(&models.StepUpdate{
				State:    models.Pointer(models.Paused),
				OldState: step.State,
				Message:  "has been paused",
			})
		}
	case "resume":
		if manager.State() == dag.Paused {
			manager.Resume()
			return storage.Task(ss.taskName).Step(ss.stepName).Update(&models.StepUpdate{
				State:    step.OldState,
				OldState: step.State,
				Message:  "has been resumed",
			})
		}
	}
	return nil
}

func (ss *StepService) Delete() error {
	storage.Task(ss.taskName).Step(ss.stepName).ClearAll()
	return nil
}

func (ss *StepService) Log() (types.Code, []*types.TaskStepLogRes, error) {
	step, err := storage.Task(ss.taskName).Step(ss.stepName).Get()
	if err != nil {
		return types.CodeFailed, nil, err
	}
	switch *step.State {
	case models.Pending:
		return types.CodePending, []*types.TaskStepLogRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   step.Message,
			},
		}, errors.New(step.Message)
	case models.Paused:
		return types.CodePaused, []*types.TaskStepLogRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   "step is paused",
			},
		}, errors.New(step.Message)
	default:
		res, _ := ss.log(nil)
		return ConvertState(*step.State), res, errors.New(step.Message)
	}
}

func (ss *StepService) log(latestLine *int64) (res []*types.TaskStepLogRes, done bool) {
	logs := storage.Task(ss.taskName).Step(ss.stepName).Log().List(latestLine)
	for _, v := range logs {
		if v.Content == common.ConsoleStart {
			continue
		}
		if v.Content == common.ConsoleDone {
			done = true
			continue
		}
		res = append(res, &types.TaskStepLogRes{
			Timestamp: v.Timestamp,
			Line:      *v.Line,
			Content:   v.Content,
		})
	}
	// 如果查询到有新日志，更新 latestLine 为最后一条日志的行号
	if len(logs) > 0 && latestLine != nil {
		*latestLine = *logs[len(logs)-1].Line // 更新 latestLine
	}
	return
}

type stateHandler func(ws *websocket.Conn, latest *int64) (bool, error)

func (ss *StepService) LogStream(ctx context.Context, ws *websocket.Conn) error {
	db := storage.Task(ss.taskName).Step(ss.stepName)
	step, err := db.Get()
	if err != nil {
		return err
	}

	var latestLine int64
	// 用于防止某些状态下的重复推送
	var onceMap = map[models.State]*sync.Once{
		models.Pending: new(sync.Once),
		models.Paused:  new(sync.Once),
		models.Unknown: new(sync.Once),
	}
	// 状态处理函数映射
	handlers := map[models.State]stateHandler{
		models.Pending: ss.createOnceHandler(onceMap[models.Pending], types.CodePending, "step is pending"),
		models.Paused:  ss.createOnceHandler(onceMap[models.Paused], types.CodePaused, "step is paused"),
		models.Unknown: ss.createOnceHandler(onceMap[models.Unknown], types.CodeNoData, "step status unknown"),
		models.Running: ss.handleRunningState,
		models.Stop:    ss.handleFinalState(types.CodeSuccess),
		models.Failed:  ss.handleFinalState(types.CodeFailed),
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// 避免多次推送
		// Paused, Pending,Unknown 状态只发送一次, 然后继续等到状态变化再继续推送
		// Running 状态会一直推送, 直到状态推送完成.
		// Stop, Failed 推送后结束.

		if handler, exists := handlers[*step.State]; exists {
			shouldContinue, err := handler(ws, &latestLine)
			if err != nil {
				logx.Errorln(err)
				return err
			}
			if !shouldContinue {
				return nil
			}
		} else {
			return errors.New("unhandled step state")
		}

		var err error
		step, err = db.Get()
		if err != nil {
			return err
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func (ss *StepService) createOnceHandler(once *sync.Once, code types.Code, message string) stateHandler {
	return func(ws *websocket.Conn, latest *int64) (bool, error) {
		once.Do(func() {
			_ = ws.WriteJSON(base.WithData([]types.TaskStepLogRes{
				{
					Timestamp: time.Now().UnixNano(),
					Line:      1,
					Content:   message,
				},
			}).WithCode(code))
		})
		return true, nil
	}
}

func (ss *StepService) handleRunningState(ws *websocket.Conn, latestLine *int64) (bool, error) {
	res, done := ss.log(latestLine)
	err := ws.WriteJSON(base.WithData(res).WithCode(types.CodeRunning).WithError(errors.New("in progress")))
	if err != nil {
		return false, err
	}
	if done {
		return false, nil
	}
	return true, nil
}

func (ss *StepService) handleFinalState(code types.Code) stateHandler {
	return func(ws *websocket.Conn, latestLine *int64) (bool, error) {
		db := storage.Task(ss.taskName).Step(ss.stepName)
		step, err := db.Get()
		if err != nil {
			return false, err
		}
		res, _ := ss.log(latestLine)
		var errMsg error
		if code == types.CodeFailed {
			errMsg = fmt.Errorf("exit code: %d", step.Code)
			if step.Message != "" {
				errMsg = fmt.Errorf(step.Message)
			}
		}
		err = ws.WriteJSON(base.WithData(res).WithCode(code).WithError(errMsg))
		if err != nil {
			return false, err
		}
		return false, nil
	}
}
