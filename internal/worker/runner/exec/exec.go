package exec

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

type SCmd struct {
	storage storage.IStep

	done      chan struct{}
	ops       cmd.Options
	cmd       *cmd.Cmd
	stderrBuf *cmd.OutputBuffer
	ctx       context.Context
	cancel    context.CancelFunc

	shell      string
	workspace  string
	scriptName string
	timeout    time.Duration
}

func New(
	storage storage.IStep,
	shell, workspace, scriptDir string,
) (*SCmd, error) {
	var c = &SCmd{
		storage:   storage,
		workspace: workspace,
		shell:     shell,
		done:      make(chan struct{}),
		ctx:       context.Background(),
		stderrBuf: cmd.NewOutputBuffer(),
	}

	c.ops = cmd.Options{
		Buffered:       true,
		Streaming:      true,
		BeforeExec:     c.beforeExec(),
		LineBufferSize: 1024,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.scriptName = filepath.Join(scriptDir, ksuid.New().String())
	c.scriptName = c.scriptName + c.scriptSuffix()
	if err := os.MkdirAll(scriptDir, os.ModePerm); err != nil {
		return nil, err
	}
	content, err := storage.Content()
	if err != nil {
		return nil, err
	}
	if c.shell == "cmd" || c.shell == "powershell" {
		content = c.utf8ToGb2312(content)
	}
	if err = os.WriteFile(c.scriptName, []byte(content), os.ModePerm); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *SCmd) scriptSuffix() string {
	switch c.shell {
	case "python", "python2", "python3", "py", "py2", "py3":
		return ".py"
	}
	return c.selfScriptSuffix()
}

func (c *SCmd) Clear() error {
	return os.Remove(c.scriptName)
}

func (c *SCmd) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		if _err := recover(); _err != nil {
			err = fmt.Errorf("%v", _err)
		}
	}()

	err = c.newCmd()
	if err != nil {
		return common.CodeSystemErr, err
	}
	// Print STDOUT and STDERR lines streaming from Cmd
	go c.consoleOutput()
	defer func() {
		c.cancel()
		<-c.done
	}()

	select {
	// 人工强制终止
	case <-ctx.Done():
		_ = c.kill()
		err = common.ErrManual
		code = common.CodeKilled
		if context.Cause(ctx) != nil {
			switch {
			case errors.Is(context.Cause(ctx), common.ErrTimeOut):
				err = common.ErrTimeOut
				code = common.CodeTimeout
			default:
				err = context.Cause(ctx)
			}
		}
	// 执行超时信号
	case <-c.ctx.Done():
		// 如果直接使用cmd.Process.Kill()并不能杀死主进程下的所有子进程
		// _ = cmd.Process.Kill()
		_ = c.kill()
		err = common.ErrTimeOut
		code = common.CodeTimeout
	// 执行结果
	case status := <-c.cmd.Start():
		code = int64(status.Exit)
		err = status.Error
		if err != nil && code == 0 {
			code = common.CodeSystemErr
		}
		if err == nil && code != 0 {
			err = fmt.Errorf("exit code %d", code)
		}
	}
	return
}

func (c *SCmd) newCmd() error {
	timeout, err := c.storage.Timeout()
	if err != nil {
		return err
	}
	if timeout > 0 {
		c.ctx, c.cancel = context.WithTimeout(c.ctx, timeout)
	}
	switch c.shell {
	case "python", "python2", "py2", "py":
		c.cmd = cmd.NewCmdOptions(c.ops, "python2", c.scriptName)
	case "python3", "py3":
		c.cmd = cmd.NewCmdOptions(c.ops, "python3", c.scriptName)
	default:
		c.selfCmd()
	}

	// 动态获取环境变量
	var envs []string
	taskEnv := c.storage.GlobalEnv().List()
	for _, env := range taskEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	stepEnv := c.storage.Env().List()
	for _, env := range stepEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}

	// inject env
	c.cmd.Env = append(os.Environ(), append(
		envs,
		fmt.Sprintf("TASK_NAME=%s", c.storage.TaskName()),
		fmt.Sprintf("TASK_STEP_NAME=%s", c.storage.Name()),
		fmt.Sprintf("TASK_WORKSPACE=%s", c.workspace),
	)...)
	// set workspace
	c.cmd.Dir = c.workspace
	return nil
}

func (c *SCmd) consoleOutput() {
	defer close(c.done)
	for {
		var line string
		var open bool
		select {
		case <-c.ctx.Done():
			if c.cmd.Stdout != nil || c.cmd.Stderr != nil {
				continue
			}
			return
		case line, open = <-c.cmd.Stdout:
			if !open {
				c.cmd.Stdout = nil
				continue
			}
		case line, open = <-c.cmd.Stderr:
			if !open {
				c.cmd.Stderr = nil
				continue
			}
			_, _ = c.stderrBuf.Write([]byte(line))
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = c.transform(line)
		c.storage.Log().Write(line)
	}
}
