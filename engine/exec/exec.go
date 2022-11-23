package exec

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xmapst/osreapi/cache"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
	Step            int64
	Content         string
	ExternalEnvVars []string
	Timeout         time.Duration
	TTL             time.Duration
	Workspace       string
	ScriptDir       string
	absFilePath     string
	exec            *exec.Cmd
	context         context.Context
	cancelFunc      context.CancelFunc
}

func (c *Cmd) Create() error {
	c.absFilePath = filepath.Join(c.ScriptDir, c.Name)
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
		if c.isGBK(line) {
			line = c.gbkToUtf8(line)
		}
		cache.SetTaskStepOutput(c.TaskID, c.Step, num, &cache.TaskStepOutput{
			Timestamp: time.Now().UnixNano(),
			Line:      num,
			Content:   string(line),
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

func (c *Cmd) isGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0x7f {
			//编码0~127,只有一个字节的编码，兼容ASCII码
			i++
			continue
		} else {
			//大于127的使用双字节编码，落在gbk编码范围内的字符
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
