package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/queues"
	"github.com/xmapst/AutoExecFlow/internal/server/api/base"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
	"github.com/xmapst/AutoExecFlow/types"
)

type SStepService struct {
	taskName string
	stepName string
}

func Step(taskName string, stepName string) *SStepService {
	return &SStepService{
		taskName: taskName,
		stepName: stepName,
	}
}

func (ss *SStepService) Create(globalTimeout time.Duration, step *types.SStepReq) error {
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
	if err = ss.saveStep(timeout, step); err != nil {
		return err
	}
	return nil
}

func (ss *SStepService) review(step *types.SStepReq) (time.Duration, error) {
	step.Name = reg.ReplaceAllString(step.Name, "")
	if step.Name == "" {
		step.Name = ksuid.New().String()
	}
	ss.stepName = step.Name

	if step.Type == "" {
		return 0, errors.New("step type is empty")
	}

	if step.Content == "" {
		return 0, errors.New("step content is empty")
	}

	// 校验env是否重复
	var envKeys []string
	for _, v := range step.Env {
		envKeys = append(envKeys, v.Name)
	}
	dup := utils.CheckDuplicate(envKeys)
	if dup != nil {
		return 0, fmt.Errorf("duplicate key %v", dup)
	}

	step.Depends = utils.RemoveDuplicate(step.Depends)
	timeout, _ := time.ParseDuration(step.Timeout)
	return timeout, nil
}

func (ss *SStepService) saveStep(timeout time.Duration, step *types.SStepReq) (err error) {
	stepStorage := storage.Task(ss.taskName).Step(step.Name)
	defer func() {
		if err != nil {
			_ = stepStorage.ClearAll()
		}
	}()
	err = storage.Task(ss.taskName).StepCreate(&models.SStep{
		TaskName: ss.taskName,
		Name:     step.Name,
		Desc:     step.Desc,
		Type:     step.Type,
		Content:  step.Content,
		Timeout:  timeout,
		Disable:  models.Pointer(step.Disable),
		SStepUpdate: models.SStepUpdate{
			Message:  "the step is waiting to be scheduled for execution",
			Code:     models.Pointer(int64(0)),
			State:    models.Pointer(models.StatePending),
			OldState: models.Pointer(models.StatePending),
		},
	})
	if err != nil {
		return fmt.Errorf("save step error: %s", err)
	}
	// save step env
	var envs models.SEnvs
	for _, env := range step.Env {
		envs = append(envs, &models.SEnv{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	if err = stepStorage.Env().Insert(envs...); err != nil {
		return fmt.Errorf("save step env error: %s", err)
	}
	// save step depend
	err = stepStorage.Depend().Insert(step.Depends...)
	if err != nil {
		return fmt.Errorf("save step depend error: %s", err)
	}
	return
}

func (ss *SStepService) Detail() (types.Code, *types.SStepRes, error) {
	stepStorage := storage.Task(ss.taskName).Step(ss.stepName)
	step, err := stepStorage.Get()
	if err != nil {
		logx.Errorln(err)
		return types.CodeFailed, nil, err
	}
	data := &types.SStepRes{
		Name:    step.Name,
		Desc:    step.Desc,
		State:   models.StateMap[*step.State],
		Code:    *step.Code,
		Message: step.Message,
		Timeout: step.Timeout.String(),
		Disable: *step.Disable,
		Type:    step.Type,
		Content: step.Content,
		Time: types.STimeRes{
			Start: step.STimeStr(),
			End:   step.ETimeStr(),
		},
	}
	data.Depends = storage.Task(ss.taskName).Step(step.Name).Depend().List()
	envs := stepStorage.Env().List()
	for _, env := range envs {
		data.Env = append(data.Env, &types.SEnv{
			Name:  env.Name,
			Value: env.Value,
		})
	}
	return types.Code(data.Code), data, nil
}

func (ss *SStepService) Manager(action string, duration string) error {
	task, err := storage.Task(ss.taskName).Get()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	if *task.State != models.StateRunning && *task.State != models.StatePending && *task.State != models.StatePaused {
		return errors.New("task is no running")
	}
	step, err := storage.Task(ss.taskName).Step(ss.stepName).Get()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	if *step.State != models.StateRunning && *step.State != models.StatePending && *step.State != models.StatePaused {
		return errors.New("step is no running")
	}
	return queues.PublishManager(task.Node, utils.JoinWithInvisibleChar(ss.taskName, ss.stepName, action, duration))
}

func (ss *SStepService) Delete() error {
	return storage.Task(ss.taskName).Step(ss.stepName).ClearAll()
}

func (ss *SStepService) Log() (types.Code, types.SStepLogsRes, error) {
	step, err := storage.Task(ss.taskName).Step(ss.stepName).Get()
	if err != nil {
		return types.CodeFailed, nil, err
	}
	switch *step.State {
	case models.StatePending:
		return types.CodePending, types.SStepLogsRes{
			{
				Timestamp: time.Now().UnixNano(),
				Line:      1,
				Content:   step.Message,
			},
		}, errors.New(step.Message)
	case models.StatePaused:
		return types.CodePaused, types.SStepLogsRes{
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

func (ss *SStepService) log(latestLine *int64) (res types.SStepLogsRes, done bool) {
	logs := storage.Task(ss.taskName).Step(ss.stepName).Log().List(latestLine)
	for _, v := range logs {
		if v.Content == common.ConsoleStart {
			continue
		}
		if v.Content == common.ConsoleDone {
			done = true
			continue
		}
		res = append(res, &types.SStepLogRes{
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

type stateHandlerFn func(ws *websocket.Conn, latest *int64) (bool, error)

func (ss *SStepService) LogStream(ctx context.Context, ws *websocket.Conn) error {
	db := storage.Task(ss.taskName).Step(ss.stepName)
	step, err := db.Get()
	if err != nil {
		return err
	}

	var latestLine int64
	// 用于防止某些状态下的重复推送
	var onceMap = map[models.State]*sync.Once{
		models.StatePending: new(sync.Once),
		models.StatePaused:  new(sync.Once),
		models.StateUnknown: new(sync.Once),
	}
	// 状态处理函数映射
	handlers := map[models.State]stateHandlerFn{
		models.StatePending: ss.createOnceHandler(onceMap[models.StatePending], types.CodePending, "step is pending"),
		models.StatePaused:  ss.createOnceHandler(onceMap[models.StatePaused], types.CodePaused, "step is paused"),
		models.StateUnknown: ss.createOnceHandler(onceMap[models.StateUnknown], types.CodeNoData, "step status unknown"),
		models.StateRunning: ss.handleRunningState,
		models.StateStopped: ss.handleFinalState(types.CodeSuccess),
		models.StateFailed:  ss.handleFinalState(types.CodeFailed),
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

func (ss *SStepService) createOnceHandler(once *sync.Once, code types.Code, message string) stateHandlerFn {
	return func(ws *websocket.Conn, latest *int64) (bool, error) {
		once.Do(func() {
			_ = ws.WriteJSON(base.WithData([]types.SStepLogRes{
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

func (ss *SStepService) handleRunningState(ws *websocket.Conn, latestLine *int64) (bool, error) {
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

func (ss *SStepService) handleFinalState(code types.Code) stateHandlerFn {
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
