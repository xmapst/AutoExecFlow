package config

import (
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

var App = new(Config)

type Config struct {
	Debug         bool
	Normal        bool
	ListenAddress string
	PoolSize      int
	ExecTimeOut   time.Duration
	WebTimeout    time.Duration
	RelativePath  string
	RootDir       string
	DBType        string
	SelfUpdateURL string
}

func (c *Config) Init() error {
	var dirs = map[string]string{
		"root":      c.RootDir,
		"script":    c.ScriptDir(),
		"data":      c.DataDir(),
		"log":       c.LogDir(),
		"workspace": c.WorkSpace(),
	}
	for name, dir := range dirs {
		if err := utils.EnsureDirExist(dir); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %v", dir, err)
		}
		logx.Infof("%s dir: %s", name, dir)
	}

	var logfile string
	if !c.Debug {
		logfile = filepath.Join(c.LogDir(), utils.ServiceName+".log")
	}
	logx.SetupLogger(logfile, zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func (c *Config) ScriptDir() string {
	return filepath.Join(c.RootDir, "scripts")
}

func (c *Config) DataDir() string {
	return filepath.Join(c.RootDir, "data")
}

func (c *Config) LogDir() string {
	return filepath.Join(c.RootDir, "logs")
}

func (c *Config) WorkSpace() string {
	return filepath.Join(c.RootDir, "workspace")
}
