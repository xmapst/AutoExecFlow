package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/segmentio/ksuid"
)

var (
	ErrTimeOut = errors.New("forced termination by timeout")
	ErrManual  = errors.New("artificial force termination")
)

const (
	Killed    int64 = -997
	Timeout   int64 = -998
	SystemErr int64 = -999
)

type Cmd struct {
	ops        cmd.Options
	cmd        *cmd.Cmd
	stderrBuf  *cmd.OutputBuffer
	ctx        context.Context
	cancel     context.CancelFunc
	env        []string
	shell      string
	content    string
	workspace  string
	scriptName string
	scriptDir  string
	timeout    time.Duration
	consoleCh  chan<- string
}

func New(options ...Option) (*Cmd, error) {
	var c = &Cmd{
		ctx:       context.Background(),
		stderrBuf: cmd.NewOutputBuffer(),
		timeout:   30 * time.Minute,
		workspace: filepath.Join(os.TempDir(), "workspace"),
		scriptDir: filepath.Join(os.TempDir(), "script"),
	}
	for _, option := range options {
		option(c)
	}
	c.ops = cmd.Options{
		Buffered:       true,
		Streaming:      true,
		BeforeExec:     c.beforeExec(),
		LineBufferSize: 1024,
	}
	c.scriptName = filepath.Join(c.scriptDir, ksuid.New().String())
	c.scriptName = c.scriptName + c.scriptSuffix()
	if err := os.MkdirAll(c.scriptDir, os.ModePerm); err != nil {
		return nil, err
	}
	if c.shell == "cmd" || c.shell == "powershell" {
		c.content = c.utf8ToGb2312(c.content)
	}
	if err := os.WriteFile(c.scriptName, []byte(c.content), os.ModePerm); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cmd) scriptSuffix() string {
	switch c.shell {
	case "python", "python2", "python3", "py", "py2", "py3":
		return ".py"
	}
	return c.selfScriptSuffix()
}

func (c *Cmd) Clear() error {
	return os.Remove(c.scriptName)
}

func (c *Cmd) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		if _err := recover(); _err != nil {
			err = fmt.Errorf("%v", _err)
		}
	}()

	c.newCmd()
	defer func() {
		c.cancel()
	}()

	// Print STDOUT and STDERR lines streaming from Cmd
	go c.consoleOutput()

	select {
	// 人工强制终止
	case <-ctx.Done():
		_ = c.kill()
		err = ErrManual
		code = Killed
		if context.Cause(ctx) != nil {
			switch {
			case errors.Is(context.Cause(ctx), ErrTimeOut):
				err = ErrTimeOut
				code = Timeout
			default:
				err = context.Cause(ctx)
			}
		}
	// 执行超时信号
	case <-c.ctx.Done():
		// 如果直接使用cmd.Process.Kill()并不能杀死主进程下的所有子进程
		// _ = cmd.Process.Kill()
		_ = c.kill()
		err = ErrTimeOut
		code = Timeout
	// 执行结果
	case status := <-c.cmd.Start():
		code = int64(status.Exit)
		err = status.Error
		if err != nil && code == 0 {
			code = SystemErr
		}
		if err == nil && code != 0 {
			err = fmt.Errorf("exit code %d%s", code, last(c.stderrBuf.Lines()))
		}
	}
	return
}

func (c *Cmd) newCmd() {
	c.ctx, c.cancel = context.WithTimeout(context.Background(), c.timeout)
	switch c.shell {
	case "python", "python2", "py2", "py":
		c.cmd = cmd.NewCmdOptions(c.ops, "python2", c.scriptName)
	case "python3", "py3":
		c.cmd = cmd.NewCmdOptions(c.ops, "python3", c.scriptName)
	default:
		c.selfCmd()
	}

	// inject env
	c.cmd.Env = append(os.Environ(), c.env...)
	// set workspace
	c.cmd.Dir = c.workspace
}

func (c *Cmd) consoleOutput() {
	defer func() {
		if c.consoleCh != nil {
			close(c.consoleCh)
		}
	}()
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

		if c.consoleCh != nil {
			c.consoleCh <- line
		}
	}
}

func last(slice []string) string {
	if len(slice) > 0 {
		return ";" + slice[len(slice)-1]
	}
	return ""
}
