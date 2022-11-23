//go:build !windows

package exec

import (
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
	}
	return ""
}

func (c *Cmd) selfCmd() bool {
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
	c.exec.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.exec.Dir = c.Workspace
	go c.run(done, errCh)
	select {
	// execute timeout signal
	case <-c.context.Done():
		msg = "exec time out"
		// If you use cmd.Process.Kill() directly, only the child process is killed,
		//but the grandchild process is not killed
		//err := cmd.Process.Kill()
		if c.exec.Process != nil {
			err := syscall.Kill(-c.exec.Process.Pid, syscall.SIGKILL)
			if err != nil {
				msg = fmt.Sprintf("%s %s", msg, err.Error())
			}
		}
		return
	// execution result output
	case <-done:
		code = 0
		if c.exec.ProcessState != nil {
			code = int64(c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
		}
		return
	// Execute exception output
	case err := <-errCh:
		if c.exec.ProcessState != nil {
			code = int64(c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
		}
		msg = err.Error()
		return
	}
}
