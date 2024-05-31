package config

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xmapst/osreapi/internal/utils"
	"github.com/xmapst/osreapi/pkg/logx"
)

var App = new(Config)

type Config struct {
	Debug         bool
	Normal        bool
	ListenAddress string
	PoolSize      int
	ExecTimeOut   time.Duration
	WebTimeout    time.Duration
	ScriptDir     string
	LogDir        string
	WorkSpace     string
	DataDir       string
	DBType        string
	SelfUpdateURL string
}

func (c *Config) Init() error {
	c.LogDir = c.dir(c.LogDir, "logs")
	c.ScriptDir = c.dir(c.ScriptDir, "scripts")
	c.WorkSpace = c.dir(c.WorkSpace, "workspace")
	c.DataDir = c.dir(c.DataDir, "data")
	if c.DBType == "" {
		c.DBType = "sqlite"
	}
	return nil
}

func (c *Config) dir(dir, sub string) string {
	dir = os.Expand(dir, func(s string) string {
		if s == "TMP" || s == "TEMP" || s == "TEMPDIR" || s == "TMPDIR" {
			return os.TempDir()
		}
		if s == "HOME" || s == "HOMEDIR" {
			_s, _ := os.UserHomeDir()
			return _s
		}
		return os.Getenv(s)
	})
	defer func() {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logx.Fatalln(err)
		}
		if sub == "logs" {
			var logfile string
			if !c.Debug {
				logfile = filepath.Join(c.LogDir, utils.ServiceName+".log")
			}
			logx.SetupLogger(logfile, zap.AddStacktrace(zapcore.ErrorLevel))
		}
		logx.Infoln(sub, "dir", dir)
	}()
	if dir != "" {
		return dir
	}
	dir = filepath.Join(utils.DefaultDir, sub)
	return dir
}
