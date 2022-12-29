//go:build windows

package exec

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os/exec"
	"syscall"
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
		c.exec = exec.CommandContext(c.context, "powershell", "-NoLogo", "-NonInteractive", "-File", c.absFilePath)
	default:
		c.exec = exec.CommandContext(c.context, "cmd", "/C", c.absFilePath)
	}
}

func (c *Cmd) Run() (code int64, msg string) {
	var done, errCh = make(chan bool), make(chan error)
	code = 255
	defer c.clear()
	c.initCmd()
	defer c.cancelFunc()
	c.exec.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	c.exec.Dir = c.Workspace
	go c.run(done, errCh, windows.GetACP())
	select {
	// 执行超时信号
	case <-c.context.Done():
		msg = "exec time out"
		// 如果直接使用cmd.Process.Kill()并不能杀死主进程下的所有子进程
		// _ = cmd.Process.Kill()
		err := KillAll(c.exec.Process.Pid)
		if err != nil {
			msg = fmt.Sprintf("%s %s", msg, err.Error())
		}
		return
	// 执行结果输出
	case <-done:
		code = 0
		if c.exec.ProcessState != nil {
			code = int64(c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
		}
		return
	// 异常输出
	case err := <-errCh:
		if c.exec.ProcessState != nil {
			code = int64(c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
		}
		msg = err.Error()
		return
	}
}
