//go:build !windows

package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/xmapst/logx"
	"golang.org/x/term"

	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

func (c *SCmd) selfScriptSuffix() string {
	switch c.shell {
	case "ash":
		return ".ash"
	case "bash":
		return ".bash"
	case "csh":
		return ".csh"
	case "dash":
		return ".dash"
	case "ksh":
		return ".ksh"
	case "shell", "sh":
		return ".sh"
	case "tcsh":
		return ".tcsh"
	case "zsh":
		return ".zsh"
	default:
		return ".bash"
	}
}

func (c *SCmd) selfCmd() *exec.Cmd {
	var cmd *exec.Cmd
	switch c.shell {
	case "ash":
		cmd = exec.CommandContext(c.ctx, "ash", "-c", c.scriptName)
	case "csh":
		cmd = exec.CommandContext(c.ctx, "csh", "-c", c.scriptName)
	case "dash":
		cmd = exec.CommandContext(c.ctx, "dash", "-c", c.scriptName)
	case "ksh":
		cmd = exec.CommandContext(c.ctx, "ksh", "-c", c.scriptName)
	case "shell", "sh":
		// 严格模式
		// c.exec = exec.CommandContext(c.ctx, "sh", "-e", c.absFilePath)
		cmd = exec.CommandContext(c.ctx, "sh", c.scriptName)
	case "tcsh":
		cmd = exec.CommandContext(c.ctx, "tcsh", "-c", c.scriptName)
	case "zsh":
		cmd = exec.CommandContext(c.ctx, "zsh", "-c", c.scriptName)
	default:
		// 严格模式
		// c.exec = exec.CommandContext(c.ctx, "bash", "--noprofile", "--norc", "-e", "-o", "pipefail", c.absFilePath)
		// -o pipefail 管道中最后一个返回非零退出状态码的命令的退出状态码将作为该管道命令的返回值，若所有命令的退出状态码都为零则返回零
		cmd = exec.CommandContext(c.ctx, "bash", "-o", "pipefail", c.scriptName)
	}
	return cmd
}

func (c *SCmd) utf8ToGb2312(s string) string {
	return s
}

func (c *SCmd) transform(line string) string {
	return line
}

func (c *SCmd) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		c.cancel()
		if _r := recover(); _r != nil {
			err = fmt.Errorf("panic during execution %v", _r)
			code = common.CodeSystemErr
			stack := debug.Stack()
			if _err, ok := _r.(error); ok && strings.Contains(_err.Error(), context.Canceled.Error()) {
				code = common.CodeKilled
				err = common.ErrManual
			}
			c.storage.Log().Write(err.Error(), string(stack))
		}
	}()
	cmd, err := c.newCmd(ctx)
	if err != nil {
		c.storage.Log().Write(err.Error())
		return common.CodeSystemErr, err
	}
	ppty, tty, err := pty.Open()
	if err != nil {
		c.storage.Log().Write(err.Error())
		return common.CodeSystemErr, err
	}

	defer func() {
		if ppty != nil {
			_ = ppty.Close()
		}
		if tty != nil {
			_ = tty.Close()
		}
	}()

	if term.IsTerminal(int(tty.Fd())) {
		_, err = term.MakeRaw(int(tty.Fd()))
		if err != nil {
			_ = ppty.Close()
			_ = tty.Close()
			c.storage.Log().Write(err.Error())
			return common.CodeSystemErr, err
		}
	}
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	r, w := io.Pipe()
	go c.copyOutput(r)
	writer := &ptyWriter{Out: w}
	logctx, finishLog := context.WithCancel(context.Background())
	go c.copyPtyOutput(writer, ppty, finishLog)
	go c.writeKeepAlive(ppty)
	err = cmd.Run()
	if cmd.ProcessState != nil {
		code = int64(cmd.ProcessState.ExitCode())
		if cmd.ProcessState.Pid() != 0 {
			_ = syscall.Kill(-cmd.ProcessState.Pid(), syscall.SIGKILL)
		}
	}
	writer.AutoStop = true
	if _, _err := tty.Write([]byte("\x04")); _err != nil {
		logx.Debugln("Failed to write EOT")
	}

	<-logctx.Done()
	if c.ctx.Err() != nil {
		switch {
		case errors.Is(context.Cause(c.ctx), common.ErrTimeOut):
			err = common.ErrTimeOut
			code = common.CodeTimeout
		default:
			err = common.ErrManual
			code = common.CodeKilled
		}
	}
	return
}

func (c *SCmd) copyPtyOutput(writer io.Writer, ppty io.Reader, finishLog context.CancelFunc) {
	defer func() {
		finishLog()
	}()
	if _, err := io.Copy(writer, ppty); err != nil {
		return
	}
}

func (c *SCmd) writeKeepAlive(ppty io.Writer) {
	n := 1
	var err error
	for n == 1 && err == nil {
		n, err = ppty.Write([]byte{4})
		<-time.After(time.Second)
	}
}

type ptyWriter struct {
	Out       io.Writer
	AutoStop  bool
	dirtyLine bool
}

// 定义正则表达式，用来匹配 ANSI 转义序列
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func (w *ptyWriter) Write(buf []byte) (int, error) {
	if w.AutoStop && len(buf) > 0 && buf[len(buf)-1] == 4 {
		n, err := w.Out.Write(buf[:len(buf)-1])
		if err != nil {
			return n, err
		}
		if w.dirtyLine || len(buf) > 1 && buf[len(buf)-2] != '\n' {
			_, _ = w.Out.Write([]byte("\n"))
			return n, io.EOF
		}
		return n, io.EOF
	}

	cleaned := ansiRegexp.ReplaceAll(buf, nil)
	var lineStart int
	for i, b := range cleaned {
		if b == '\r' || b == '\n' {
			if i > lineStart {
				_, err := w.Out.Write(cleaned[lineStart:i])
				if err != nil {
					return len(buf), err
				}
			}
			if b == '\n' || i == len(cleaned)-1 {
				_, err := w.Out.Write([]byte("\n"))
				if err != nil {
					return len(buf), err
				}
			}
			lineStart = i + 1
		}
	}
	w.dirtyLine = strings.LastIndex(string(buf), "\n") < len(buf)-1
	return len(buf), nil
}
