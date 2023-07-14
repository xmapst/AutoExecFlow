package exec

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/logx"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	ErrTimeOut = errors.New("forced termination by timeout")
	ErrManual  = errors.New("artificial force termination")
)

type Cmd struct {
	log             logx.Logger
	Shell           string
	Name            string
	TaskID          string
	StepID          int64
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
	c.log = logx.GetSubLoggerWithKeyValue(map[string]string{
		"task":       c.TaskID,
		"workspace":  c.Workspace,
		"step_id":    strconv.FormatInt(c.StepID, 10),
		"step_name":  c.Name,
		"step_shell": c.Shell,
	}).GetSubLoggerWithOption(zap.AddCallerSkip(-1))
	if err := os.MkdirAll(c.ScriptDir, os.ModePerm); err != nil {
		c.log.Errorln(err)
		return err
	}
	c.absFilePath = filepath.Join(c.ScriptDir, c.Name)
	c.absFilePath = c.absFilePath + c.scriptSuffix()
	c.log.Infof("create script %s", filepath.Base(c.absFilePath))
	if c.Shell == "cmd" || c.Shell == "powershell" {
		c.Content = c.utf8ToGb2312(c.Content)
	}
	if err := os.WriteFile(c.absFilePath, []byte(c.Content), os.ModePerm); err != nil {
		c.log.Errorln(err)
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
	c.log.Infof("cleanup script %s", filepath.Base(c.absFilePath))
	err := os.Remove(c.absFilePath)
	if err != nil {
		c.log.Errorln(err)
	}
}

func (c *Cmd) initCmd(ctx context.Context) {
	c.context, c.cancelFunc = context.WithTimeout(ctx, c.Timeout)
	c.commonCmd()
	c.injectionEnv()
}

func (c *Cmd) commonCmd() {
	switch c.Shell {
	case "python", "python2", "py2", "py":
		c.exec = exec.CommandContext(c.context, "python2", c.absFilePath)
	case "python3", "py3":
		c.exec = exec.CommandContext(c.context, "python3", c.absFilePath)
	default:
		c.selfCmd()
	}
}

func (c *Cmd) injectionEnv() {
	c.log.Infoln("inject customize env", c.ExternalEnvVars)
	c.exec.Env = append(append(os.Environ(), c.ExternalEnvVars...),
		fmt.Sprintf("WRE_SELF_UPDATE_TASK_ID=%s", c.TaskID),
		fmt.Sprintf("TASK_ID=%s", c.TaskID),
		fmt.Sprintf("TASK_STEP_ID=%d", c.StepID),
	)
}

func (c *Cmd) run(done chan bool, errCh chan error, acp uint32) {
	stdout, err := c.exec.StdoutPipe()
	if err != nil {
		c.log.Errorln(err)
		errCh <- err
		return
	}
	c.exec.Stderr = c.exec.Stdout

	// 实时写入缓存及落盘
	go c.output(stdout, acp)

	err = c.exec.Run()
	if err != nil {
		c.log.Errorln(err)
		errCh <- err
		return
	}
	done <- true
}

func (c *Cmd) output(stdout io.ReadCloser, acp uint32) {
	reader := bufio.NewReader(stdout)
	var num int64 = 1
	for {
		line, _, err := reader.ReadLine()
		if err != nil || err == io.EOF {
			c.log.Errorln(err)
			break
		}
		line = bytes.TrimSpace(line)
		if line == nil {
			continue
		}
		if c.isGBK(line) || acp == 936 {
			line = c.gbkToUtf8(line)
		}
		cache.SetTaskStepOutput(c.TaskID, c.StepID, num, &cache.TaskStepOutput{
			Timestamp: time.Now().UnixNano(),
			Line:      num,
			Content:   string(line),
		}, c.TTL+c.Timeout)
		c.log.Infoln(string(line))
		num += 1
	}
	cache.SetTaskStepOutputDone(c.TaskID, c.StepID, num, c.TTL+c.Timeout)
}

func (c *Cmd) gbkToUtf8(s []byte) []byte {
	defer func() {
		err := recover()
		if err != nil {
			c.log.Errorln(err)
		}
	}()
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	b, err := io.ReadAll(reader)
	if err != nil {
		c.log.Errorln(err)
		return s
	}
	return b
}

func (c *Cmd) utf8ToGb2312(s string) string {
	defer func() {
		err := recover()
		if err != nil {
			c.log.Errorln(err)
		}
	}()
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		c.log.Errorln(err)
		return s
	}

	return string(d)
}

func (c *Cmd) isGBK(data []byte) bool {
	defer func() {
		err := recover()
		if err != nil {
			c.log.Errorln(err)
		}
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
