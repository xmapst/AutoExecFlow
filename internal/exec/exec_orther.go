//go:build !windows

package exec

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

func (c *Cmd) selfScriptSuffix() string {
	switch c.Shell {
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

func (c *Cmd) selfCmd() {
	switch c.Shell {
	case "ash":
		c.exec = exec.CommandContext(c.context, "ash", "-c", c.absFilePath)
	case "bash":
		c.exec = exec.CommandContext(c.context, "bash", "-c", c.absFilePath)
	case "csh":
		c.exec = exec.CommandContext(c.context, "csh", "-c", c.absFilePath)
	case "dash":
		c.exec = exec.CommandContext(c.context, "dash", "-c", c.absFilePath)
	case "ksh":
		c.exec = exec.CommandContext(c.context, "ksh", "-c", c.absFilePath)
	case "shell", "sh":
		c.exec = exec.CommandContext(c.context, "sh", "-c", c.absFilePath)
	case "tcsh":
		c.exec = exec.CommandContext(c.context, "tcsh", "-c", c.absFilePath)
	case "zsh":
		c.exec = exec.CommandContext(c.context, "zsh", "-c", c.absFilePath)
	default:
		c.exec = exec.CommandContext(c.context, "bash", "-c", c.absFilePath)
	}
}

func (c *Cmd) Run(ctx context.Context) (code int64, err error) {
	defer func() {
		_err := recover()
		if _err != nil {
			err = fmt.Errorf("%v", _err)
		}
	}()
	var done, errCh = make(chan bool), make(chan error)
	code = 255
	defer c.clear()
	c.initCmd(ctx)
	defer c.cancelFunc()
	c.exec.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.exec.Dir = c.Workspace
	go c.run(done, errCh, 0)
	select {
	// 人工强制终止
	case <-ctx.Done():
		if c.exec != nil && c.exec.Process != nil {
			_ = syscall.Kill(-c.exec.Process.Pid, syscall.SIGKILL)
		}
		err = ErrManual
	// 执行超时信号
	case <-c.context.Done():
		// err = errors.New("exec time out")
		// If you use cmd.Process.Kill() directly, only the child process is killed,
		// but the grandchild process is not killed
		// err := cmd.Process.Kill()
		if c.exec != nil && c.exec.Process != nil {
			err = syscall.Kill(-c.exec.Process.Pid, syscall.SIGKILL)
		}
		if err == nil {
			err = ErrTimeOut
		}
	// 执行成功
	case <-done:
		code = 0
		if c.exec.ProcessState != nil {
			code = int64(c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
		}
	// 执行异常
	case err = <-errCh:
		if c.exec.ProcessState != nil {
			code = int64(c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
		}
	}
	return
}
