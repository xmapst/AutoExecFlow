package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xmapst/osreapi/internal/logx"
)

var App = new(Config)

type Config struct {
	Debug          bool
	Normal         bool
	ListenAddress  string
	MaxRequests    int64
	PoolSize       int
	KeyExpire      time.Duration
	ExecTimeOut    time.Duration
	WebTimeout     time.Duration
	ServiceName    string
	SelfUpdateData string
	RootDir        string
	ScriptDir      string
	LogDir         string
	WorkSpace      string
	DataDir        string
}

func init() {
	executable, err := os.Executable()
	if err != nil {
		logx.Fatalln(err)
	}
	App.ServiceName = strings.TrimSuffix(filepath.Base(executable), ".exe")
	App.SelfUpdateData = filepath.Join(filepath.Dir(executable), App.ServiceName+".dat")
}

func (c *Config) Load() error {
	_ = os.MkdirAll(c.RootDir, 0777)
	logx.Infoln("root dir", c.RootDir)

	c.ScriptDir = filepath.Join(c.RootDir, "scripts")
	_ = os.MkdirAll(c.ScriptDir, 0777)
	logx.Infoln("scripts dir", c.ScriptDir)

	c.LogDir = filepath.Join(c.RootDir, "logs")
	_ = os.MkdirAll(c.LogDir, 0777)
	logx.Infoln("logs dir", c.ScriptDir)

	c.WorkSpace = filepath.Join(c.RootDir, "workspace")
	_ = os.MkdirAll(c.WorkSpace, 0777)
	logx.Infoln("workspace dir", c.ScriptDir)

	c.DataDir = filepath.Join(c.RootDir, "data")
	_ = os.MkdirAll(c.RootDir, 0777)
	logx.Infoln("data dir", c.ScriptDir)

	return nil
}
