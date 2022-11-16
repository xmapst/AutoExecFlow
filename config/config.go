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
	TempDir        string
	ScriptDir      string
	LogDir         string
}

func (c *Config) Load() error {
	executable, err := os.Executable()
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	c.ServiceName = strings.TrimSuffix(filepath.Base(executable), ".exe")
	c.SelfUpdateData = filepath.Join(filepath.Dir(executable), c.ServiceName+".dat")
	c.TempDir = filepath.Join(os.TempDir(), c.ServiceName)
	c.ScriptDir = filepath.Join(c.TempDir, "scripts")
	c.LogDir = filepath.Join(c.TempDir, "logs")
	return nil
}
