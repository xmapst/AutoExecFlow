package engine

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/config"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type Cmd struct {
	Log             *logrus.Entry
	Shell           string
	Name            string
	TaskID          string
	Step            int64
	Content         string
	ExternalEnvVars []string
	Timeout         time.Duration
	TTL             time.Duration
	workSpace       string
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
	err := os.WriteFile(c.absFilePath, []byte(c.Content), 0777)
	if err != nil {
		return err
	}
	c.workSpace = filepath.Join(config.App.WorkSpace, c.TaskID)
	err = os.MkdirAll(c.workSpace, 0777)
	if err != nil && err != os.ErrExist {
		return err
	}
	return nil
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
	_ = os.RemoveAll(c.workSpace)
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

func (c *Cmd) run(done chan bool, errCh chan error) {
	c.exec.Stderr = c.exec.Stdout
	stdout, err := c.exec.StdoutPipe()
	if err != nil {
		errCh <- err
		return
	}
	// 实时写入缓存及落盘
	go c.output(stdout)

	err = c.exec.Run()
	if err != nil {
		errCh <- err
		return
	}
	done <- true
}

func (c *Cmd) output(stdout io.ReadCloser) {
	reader := bufio.NewReader(stdout)
	var num int64 = 1
	for {
		line, _, err := reader.ReadLine()
		if err != nil || err == io.EOF {
			break
		}
		line = bytes.TrimSpace(line)
		if line == nil {
			continue
		}
		if runtime.GOOS == "windows" {
			// windows 输出转码
			line = c.gbkToUtf8(line)
		}
		cache.SetTaskStepOutput(c.TaskID, c.Step, num, &cache.TaskStepOutput{
			Line:    num,
			Content: string(line),
		}, c.TTL+c.Timeout)
		c.Log.Println(string(line))
		num += 1
	}
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
