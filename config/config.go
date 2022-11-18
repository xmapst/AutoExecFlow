package config

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var App = new(Config)

type Config struct {
	Debug          bool
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
}

func init() {
	executable, err := os.Executable()
	if err != nil {
		logrus.Fatalln(err)
	}
	App.ServiceName = strings.TrimSuffix(filepath.Base(executable), ".exe")
	App.SelfUpdateData = filepath.Join(filepath.Dir(executable), App.ServiceName+".dat")
}

func (c *Config) Load() error {
	_ = os.MkdirAll(c.RootDir, 0777)
	c.ScriptDir = filepath.Join(c.RootDir, "scripts")
	_ = os.MkdirAll(c.ScriptDir, 0777)
	c.LogDir = filepath.Join(c.RootDir, "logs")
	_ = os.MkdirAll(c.LogDir, 0777)
	c.WorkSpace = filepath.Join(c.RootDir, "workspace")
	_ = os.MkdirAll(c.WorkSpace, 0777)
	return nil
}
