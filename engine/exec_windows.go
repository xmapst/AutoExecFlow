//go:build windows

package engine

import (
	"fmt"
	"os/exec"
	"syscall"
)

func (c *Cmd) selfScriptSuffix() string {
	switch c.Shell {
	case "cmd", "bat":
		return ".bat"
	case "powershell", "ps", "ps1":
		return ".ps1"
	}
	return ""
}

func (c *Cmd) selfCmd() bool {
	switch c.Shell {
	case "cmd", "bat":
		c.exec = exec.CommandContext(c.context, "cmd", "/C", c.absFilePath)
	case "powershell", "ps", "ps1":
		c.exec = exec.CommandContext(c.context, "powershell", "-NoLogo", "-NonInteractive", "-File", c.absFilePath)
	}
	return c.exec != nil
}

func (c *Cmd) Run() (code int64, msg string) {
	var done, errCh = make(chan bool), make(chan error)
	code = 255
	defer c.clear()
	if !c.initCmd() {
		msg = "command type not found"
		return
	}
	defer c.cancelFunc()
	c.exec.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	c.exec.Dir = c.workSpace
	go c.run(done, errCh)
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
