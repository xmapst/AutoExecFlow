//go:build !windows

package engine

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

func (c *Cmd) Run() (exitCode int, content string) {
	var outputCh, errChan = make(chan string), make(chan error)
	defer c.clear()
	if !c.initCmd() {
		return 255, "command type not found"
	}
	defer c.cancelFunc()
	c.exec.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	//cmd.Dir, _ = os.UserHomeDir()
	go c.combinedOutput(outputCh, errChan)
	select {
	// execute timeout signal
	case <-c.context.Done():
		// If you use cmd.Process.Kill() directly, only the child process is killed,
		//but the grandchild process is not killed
		//err := cmd.Process.Kill()
		if c.exec.Process != nil {
			_ = syscall.Kill(-c.exec.Process.Pid, syscall.SIGKILL)
		}
		return 255, "exec time out"
	// execution result output
	case output := <-outputCh:
		code := 0
		if c.exec.ProcessState != nil {
			code = c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		}
		return code, output
	// Execute exception output
	case err := <-errChan:
		code := 255
		if c.exec.ProcessState != nil {
			code = c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		}
		return code, err.Error()
	}
}

func (c *Cmd) combinedOutput(outputCh chan string, errCh chan error) {
	output, err := c.exec.CombinedOutput()
	if err != nil {
		errCh <- fmt.Errorf(string(output) + err.Error())
		return
	}
	go c.printOutput(output)
	outputCh <- string(output)
}
