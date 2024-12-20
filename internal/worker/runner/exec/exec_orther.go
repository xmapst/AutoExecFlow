//go:build !windows

package exec

import (
	"os/exec"
	"syscall"

	"github.com/go-cmd/cmd"
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

func (c *SCmd) beforeExec() []func(cmd *exec.Cmd) {
	return []func(cmd *exec.Cmd){
		func(cmd *exec.Cmd) {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
			}
		},
	}
}

func (c *SCmd) selfCmd() {
	switch c.shell {
	case "ash":
		c.cmd = cmd.NewCmdOptions(c.ops, "ash", "-c", c.scriptName)
	case "csh":
		c.cmd = cmd.NewCmdOptions(c.ops, "csh", "-c", c.scriptName)
	case "dash":
		c.cmd = cmd.NewCmdOptions(c.ops, "dash", "-c", c.scriptName)
	case "ksh":
		c.cmd = cmd.NewCmdOptions(c.ops, "ksh", "-c", c.scriptName)
	case "shell", "sh":
		// 严格模式
		// c.exec = exec.CommandContext(c.ctx, "sh", "-e", c.absFilePath)
		c.cmd = cmd.NewCmdOptions(c.ops, "sh", c.scriptName)
	case "tcsh":
		c.cmd = cmd.NewCmdOptions(c.ops, "tcsh", "-c", c.scriptName)
	case "zsh":
		c.cmd = cmd.NewCmdOptions(c.ops, "zsh", "-c", c.scriptName)
	default:
		// 严格模式
		// c.exec = exec.CommandContext(c.ctx, "bash", "--noprofile", "--norc", "-e", "-o", "pipefail", c.absFilePath)
		// -o pipefail 管道中最后一个返回非零退出状态码的命令的退出状态码将作为该管道命令的返回值，若所有命令的退出状态码都为零则返回零
		c.cmd = cmd.NewCmdOptions(c.ops, "bash", "-o", "pipefail", c.scriptName)
	}
}

func (c *SCmd) kill() (err error) {
	if c.cmd == nil {
		return
	}
	return c.cmd.Stop()
}

func (c *SCmd) utf8ToGb2312(s string) string {
	return s
}

func (c *SCmd) transform(line string) string {
	return line
}
