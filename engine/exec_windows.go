//go:build windows

package engine

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
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

func (c *Cmd) Run() (exitCode int, content string) {
	var outputCh, errChan = make(chan string), make(chan error)
	defer c.clear()
	if !c.initCmd() {
		return 255, "command type not found"
	}
	defer c.cancelFunc()
	c.exec.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Dir, _ = os.UserHomeDir()
	go c.combinedOutput(outputCh, errChan)
	select {
	// 执行超时信号
	case <-c.context.Done():
		// 如果直接使用cmd.Process.Kill()并不能杀死主进程下的所有子进程
		// _ = cmd.Process.Kill()
		err := KillAll(c.exec.Process.Pid)
		msg := "exec time out"
		if err != nil {
			msg += err.Error()
		}
		return 255, msg
	// 执行结果输出
	case output := <-outputCh:
		var code = 0
		if c.exec.ProcessState != nil {
			code = c.exec.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		}
		return code, output
	// 异常输出
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
	// windows 输出转码 TODO:转码暂时有问题, 待做
	output = c.gbkToUtf8(output)
	go c.printOutput(output)
	outputCh <- string(output)
}

func (c *Cmd) gbkToUtf8(s []byte) []byte {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		logrus.Error(err)
		return s
	}
	return b
}
