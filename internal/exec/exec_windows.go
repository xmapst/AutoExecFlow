//go:build windows

package exec

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

func (c *Cmd) selfScriptSuffix() string {
	switch c.Shell {
	case "cmd", "bat":
		return ".bat"
	case "powershell", "ps", "ps1":
		return ".ps1"
	default:
		return ".bat"
	}
}

func (c *Cmd) selfCmd() {
	switch c.Shell {
	case "cmd", "bat":
		c.exec = exec.CommandContext(c.context, "cmd", "/C", c.absFilePath)
	case "powershell", "ps", "ps1":
		// 解决用户不写exit时, powershell进程外获取不到退出码
		command := fmt.Sprintf("%s;exit $LASTEXITCODE", c.absFilePath)
		c.exec = exec.CommandContext(c.context, "powershell", "-NoLogo", "-NonInteractive", "-Command", command)
	default:
		c.exec = exec.CommandContext(c.context, "cmd", "/C", c.absFilePath)
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
	c.exec.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	c.exec.Dir = c.Workspace
	go c.run(done, errCh, windows.GetACP())
	select {
	// 人工强制终止
	case <-ctx.Done():
		err = ErrManual
	// 执行超时信号
	case <-c.context.Done():
		// 如果直接使用cmd.Process.Kill()并不能杀死主进程下的所有子进程
		// _ = cmd.Process.Kill()
		err = KillAll(c.exec.Process.Pid)
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
