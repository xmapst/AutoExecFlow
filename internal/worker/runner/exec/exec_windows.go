//go:build windows

package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-cmd/cmd"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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

func (c *SCmd) beforeExec() []func(cmd *exec.Cmd) {
	return []func(cmd *exec.Cmd){
		func(cmd *exec.Cmd) {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				HideWindow: true,
			}
		},
	}
}

func (c *SCmd) selfCmd() {
	switch c.shell {
	case "cmd", "bat":
		c.cmd = cmd.NewCmdOptions(c.ops, "cmd", "/D", "/E:ON", "/V:OFF", "/Q", "/S", "/C", c.scriptName)
	case "powershell", "ps", "ps1":
		// 解决用户不写exit时, powershell进程外获取不到退出码
		command := fmt.Sprintf("$ErrorActionPreference='Continue';%s;exit $LASTEXITCODE", c.scriptName)
		// 激进方式, 强制用户脚本没问题
		// command := fmt.Sprintf("$ErrorActionPreference='Stop';%s;exit $LASTEXITCODE", c.absFilePath)
		c.cmd = cmd.NewCmdOptions(c.ops, "powershell", "-NoLogo", "-NonInteractive", "-Command", command)
	default:
		c.cmd = cmd.NewCmdOptions(c.ops, "cmd", "/D", "/E:ON", "/V:OFF", "/Q", "/S", "/C", c.scriptName)
	}
}

func (c *SCmd) kill() error {
	if c.cmd.Status().PID == 0 {
		return nil
	}
	kill := exec.Command("TASKKILL.exe", "/T", "/F", "/PID", strconv.Itoa(c.cmd.Status().PID))
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
