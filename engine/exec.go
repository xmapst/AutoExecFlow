package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/config"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Cmd struct {
	Log             *logrus.Entry
	Shell           string
	Name            string
	TaskID          string
	Step            int
	Content         string
	ExternalEnvVars []string
	Timeout         time.Duration
	absFilePath     string
	exec            *exec.Cmd
	context         context.Context
	cancelFunc      context.CancelFunc
}

func (c *Cmd) Create() error {
	c.absFilePath = filepath.Join(config.App.ScriptDir, c.Name)
	suffix := c.scriptSuffix()
	if suffix == "" {
		return fmt.Errorf("wrong script type")
	}
	c.absFilePath = c.absFilePath + suffix
	c.Log.Infof("create script %s", filepath.Base(c.absFilePath))
	return os.WriteFile(c.absFilePath, []byte(c.Content), 0777)
}

func (c *Cmd) scriptSuffix() string {
	switch c.Shell {
	case "python", "python2", "python3", "py", "py2", "py3":
		return ".py"
	}
	return c.selfScriptSuffix()
}

func (c *Cmd) clear() {
	// clear tmp script
	c.Log.Infof("cleanup script %s", filepath.Base(c.absFilePath))
	_ = os.Remove(c.absFilePath)
}

func (c *Cmd) initCmd() bool {
	c.context, c.cancelFunc = context.WithTimeout(context.Background(), c.Timeout)
	if c.commonCmd() || c.selfCmd() {
		c.injectionEnv()
		return true
	}
	return false
}

func (c *Cmd) commonCmd() bool {
	switch c.Shell {
	case "python", "python2", "py2", "py":
		c.exec = exec.CommandContext(c.context, "python2", c.absFilePath)
	case "python3", "py3":
		c.exec = exec.CommandContext(c.context, "python3", c.absFilePath)
	}
	return c.exec != nil
}

func (c *Cmd) injectionEnv() {
	c.exec.Env = append(append(os.Environ(), c.ExternalEnvVars...), fmt.Sprintf("WRE_SELF_UPDATE_TASK_ID=%s", c.TaskID))
}

func (c *Cmd) printOutput(output []byte) {
	reader := bufio.NewReader(bytes.NewReader(output))
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if io.EOF == err {
				break
			}
			c.Log.Errorln(err)
			continue
		}
		if len(line) != 0 {
			c.Log.Infoln(string(line))
		}
	}
}
