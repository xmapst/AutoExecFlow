//go:build windows

package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

var acp = windows.GetACP()

func (c *SCmd) selfScriptSuffix() string {
	switch c.shell {
	case "cmd", "bat":
		return ".bat"
	case "powershell", "ps", "ps1":
		return ".ps1"
	default:
		return ".bat"
	}
}

func (c *SCmd) selfCmd() *exec.Cmd {
	var cmd *exec.Cmd
	switch c.shell {
	case "cmd", "bat":
		cmd = exec.CommandContext(c.ctx, "cmd", "/D", "/E:ON", "/V:OFF", "/Q", "/S", "/C", c.scriptName)
	case "powershell", "ps", "ps1":
		// 解决用户不写exit时, powershell进程外获取不到退出码
		command := fmt.Sprintf("$ErrorActionPreference='Continue';%s;exit $LASTEXITCODE", c.scriptName)
		// 激进方式, 强制用户脚本没问题
		// command := fmt.Sprintf("$ErrorActionPreference='Stop';%s;exit $LASTEXITCODE", c.absFilePath)
		cmd = exec.CommandContext(c.ctx, "powershell", "-NoLogo", "-NonInteractive", "-Command", command)
	default:
		cmd = exec.CommandContext(c.ctx, "cmd", "/D", "/E:ON", "/V:OFF", "/Q", "/S", "/C", c.scriptName)
	}
	return cmd
}

func (c *SCmd) Run(ctx context.Context) (exit int64, err error) {
	defer func() {
		c.cancel()
		if _r := recover(); _r != nil {
			err = fmt.Errorf("panic during execution %v", _r)
			exit = common.CodeSystemErr
			stack := debug.Stack()
			if _err, ok := _r.(error); ok && strings.Contains(_err.Error(), context.Canceled.Error()) {
				exit = common.CodeKilled
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
	// 设置输出

	reader, err := cmd.StdoutPipe()
	if err != nil {
		c.storage.Log().Write(err.Error())
		return common.CodeSystemErr, err
	}
	cmd.Stderr = cmd.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
	go c.copyOutput(reader)
	err = cmd.Run()
	if cmd.ProcessState != nil {
		exit = int64(cmd.ProcessState.ExitCode())
		if cmd.ProcessState.Pid() != 0 {
			_ = c.kill(cmd.ProcessState.Pid())
		}
	}
	if err != nil && exit == 0 {
		exit = common.CodeFailed
	}
	if c.ctx.Err() != nil {
		switch {
		case errors.Is(context.Cause(c.ctx), common.ErrTimeOut):
			err = common.ErrTimeOut
			exit = common.CodeTimeout
		default:
			err = common.ErrManual
			exit = common.CodeKilled
		}
	}
	return
}

func (c *SCmd) kill(pid int) error {
	if pid == 0 {
		return nil
	}
	kill := exec.Command("TASKKILL.exe", "/T", "/F", "/PID", strconv.Itoa(pid))
	return kill.Run()
}

func (c *SCmd) transform(line string) string {
	if c.isGBK(line) || acp == 936 {
		line = string(c.gbkToUtf8([]byte(line)))
	}
	return line
}

func (c *SCmd) gbkToUtf8(s []byte) []byte {
	defer func() {
		recover()
	}()
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		return s
	}
	return b
}

func (c *SCmd) utf8ToGb2312(s string) string {
	defer func() {
		recover()
	}()
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		return s
	}

	return string(d)
}

func (c *SCmd) isGBK(data string) bool {
	defer func() {
		recover()
	}()
	length := len(data)
	var i = 0
	for i < length {
		if data[i] <= 0x7f {
			// 编码0~127,只有一个字节的编码，兼容ASCII码
			i++
			continue
		} else {
			// 大于127的使用双字节编码，落在gbk编码范围内的字符
			if data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i+1] >= 0x40 &&
				data[i+1] <= 0xfe &&
				data[i+1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}
