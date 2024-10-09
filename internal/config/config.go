package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xmapst/AutoExecFlow/internal/queues"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

var (
	App           = new(Config)
	TaskQueueName = fmt.Sprintf("%s_Worker_%s", utils.ServiceName, utils.HostName())
)

const (
	RUN_MODE_Standalone = "standalone"
	RUN_MODE_Worker     = "worker"
	RUN_MODE_Api        = "api"
)

type Config struct {
	Mode          string
	ListenAddress string
	PoolSize      int
	ExecTimeOut   time.Duration
	RelativePath  string
	RootDir       string
	Database      string
	Queue         string
	SelfUpdateURL string
	LogOutput     string
	LogLevel      string
}

func (c *Config) Init() error {
	var logfile string
	if c.LogOutput == "file" {
		logfile = filepath.Join(c.LogDir(), utils.ServiceName+".log")
	}
	logx.SetupLogger(logfile, zap.AddStacktrace(zapcore.ErrorLevel))
	level, err := zapcore.ParseLevel(c.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}
	logx.SetLevel(level)
	logx.Infof("run mode: %s", c.Mode)
	if c.Mode != RUN_MODE_Standalone {
		// 不能使用内存队列
		if strings.HasPrefix(c.Queue, queues.BROKER_INMEMORY) {
			return fmt.Errorf("can not use memory queue in worker or api mode")
		}
		if strings.HasPrefix(c.Database, storage.TYPE_SQLITE) {
			logx.Warnln("sqlite is not recommended for production use")
		}
	}

	before, _, found := strings.Cut(c.Database, "://")
	if !found {
		return fmt.Errorf("invalid database url")
	}
	if before == storage.TYPE_SQLITE {
		dir := filepath.Join(c.RootDir, "data")
		if err = utils.EnsureDirExist(dir); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %v", dir, err)
		}
		file := filepath.Join(dir, fmt.Sprintf("%s.db3", utils.ServiceName))
		logx.Infof("%s file: %s", "data", file)
		c.Database = fmt.Sprintf("%s://%s", storage.TYPE_SQLITE, file)
	}

	var dirs = map[string]string{
		"root":      c.RootDir,
		"script":    c.ScriptDir(),
		"log":       c.LogDir(),
		"workspace": c.WorkSpace(),
	}
	for name, dir := range dirs {
		if c.Mode == RUN_MODE_Api && name == "script" {
			continue
		}
		if c.Mode == RUN_MODE_Api && name == "workspace" {
			continue
		}
		if err = utils.EnsureDirExist(dir); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %v", dir, err)
		}
		logx.Infof("%s dir: %s", name, dir)
	}

	// setup queue
	err = queues.New(c.Queue)
	if err != nil {
		return fmt.Errorf("failed to setup queue: %v", err)
	}
	return nil
}

func (c *Config) ScriptDir() string {
	return filepath.Join(c.RootDir, "scripts")
}

func (c *Config) LogDir() string {
	return filepath.Join(c.RootDir, "logs")
}

func (c *Config) WorkSpace() string {
	return filepath.Join(c.RootDir, "workspace")
}
