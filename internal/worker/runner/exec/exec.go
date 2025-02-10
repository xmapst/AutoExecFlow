package exec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

type SCmd struct {
	storage storage.IStep

	ctx     context.Context
	cancel  context.CancelFunc
	envPath string

	shell      string
	workspace  string
	scriptName string
	timeout    time.Duration
}

func New(storage storage.IStep,
	shell, workspace, scriptDir string,
) (*SCmd, error) {
	var c = &SCmd{
		storage:   storage,
		workspace: workspace,
		shell:     shell,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.scriptName = filepath.Join(scriptDir, ksuid.New().String())
	c.scriptName = c.scriptName + c.scriptSuffix()
	if err := os.MkdirAll(scriptDir, os.ModePerm); err != nil {
		return nil, err
	}
	content, err := storage.Content()
	if err != nil {
		return nil, err
	}
	if c.shell == "cmd" || c.shell == "powershell" {
		content = c.utf8ToGb2312(content)
	}
	if err = os.WriteFile(c.scriptName, []byte(content), os.ModePerm); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *SCmd) scriptSuffix() string {
	switch c.shell {
	case "python", "python2", "python3", "py", "py2", "py3":
		return ".py"
	}
	return c.selfScriptSuffix()
}

func (c *SCmd) Clear() error {
	return os.Remove(c.scriptName)
}

func (c *SCmd) envs() []string {
	var envs []string
	taskEnv := c.storage.GlobalEnv().List()
	for _, env := range taskEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}
	stepEnv := c.storage.Env().List()
	for _, env := range stepEnv {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Name, env.Value))
	}

	return append(os.Environ(), append(
		envs,
		fmt.Sprintf("TASK_NAME=%s", c.storage.TaskName()),
		fmt.Sprintf("TASK_STEP_NAME=%s", c.storage.Name()),
		fmt.Sprintf("TASK_WORKSPACE=%s", c.workspace),
	)...)
}

func (c *SCmd) parseEnvFileFromFile() {
	// 打开源文件
	file, err := os.Open(c.envPath)
	if err != nil {
		logx.Warnln(err)
		return
	}
	defer file.Close()
	s := bufio.NewScanner(file)
	firstLine := true
	for s.Scan() {
		line := s.Text()

		if firstLine {
			firstLine = false
			// skip utf8 bom, powershell 5 legacy uses it for utf8
			if len(line) >= 3 && line[0] == 239 && line[1] == 187 && line[2] == 191 {
				line = line[3:]
			}
		}

		// 处理单行和多行环境变量
		singleLineEnv := strings.Index(line, "=")
		multiLineEnv := strings.Index(line, "<<")
		if singleLineEnv != -1 && (multiLineEnv == -1 || singleLineEnv < multiLineEnv) {
			// TODO: write to storage
			logx.Debugf("parsed env: %v=%v", line[:singleLineEnv], line[singleLineEnv+1:])
		} else if multiLineEnv != -1 {
			multiLineEnvContent := ""
			multiLineEnvDelimiter := line[multiLineEnv+2:]
			delimiterFound := false
			for s.Scan() {
				content := s.Text()
				if content == multiLineEnvDelimiter {
					delimiterFound = true
					break
				}
				if multiLineEnvContent != "" {
					multiLineEnvContent += "\n"
				}
				multiLineEnvContent += content
			}
			if !delimiterFound {
				logx.Errorf("invalid format delimiter '%v' not found before end of file", multiLineEnvDelimiter)
				return
			}
			// TODO: write to storage
			logx.Debugf("parsed env: %v=%v", line[:multiLineEnv], multiLineEnvContent)
		} else {
			logx.Errorf("invalid format '%v', expected a line with '=' or '<<'", line)
			return
		}
	}

	if err = s.Err(); err != nil {
		logx.Errorf("error reading file: %v", err)
		return
	}
}

func (c *SCmd) newCmd(ctx context.Context) (*exec.Cmd, error) {
	timeout, err := c.storage.Timeout()
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		c.ctx, c.cancel = context.WithTimeoutCause(ctx, timeout, common.ErrTimeOut)
	}
	var cmd *exec.Cmd
	switch c.shell {
	case "python", "python2", "py2", "py":
		cmd = exec.CommandContext(c.ctx, "python2", c.scriptName)
	case "python3", "py3":
		cmd = exec.CommandContext(c.ctx, "python3", c.scriptName)
	default:
		cmd = c.selfCmd()
	}
	cmd.Dir = c.workspace
	cmd.Env = c.envs()
	return cmd, nil
}

func (c *SCmd) copyOutput(reader io.ReadCloser) {
	defer func() {
		_ = reader.Close()
	}()
	// 按行读取输出写入到日志中
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = c.transform(line)
		c.storage.Log().Write(line)
	}
}
