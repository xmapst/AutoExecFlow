package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/xmapst/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/utils"
)

var App = &SConfig{
	DBUrl: "sqlite:///tmp/sqlite.db3",
	MQUrl: "inmemory://localhost",
}

type SConfig struct {
	Address       string        `mapstructure:"ADDR"`
	PoolSize      int           `mapstructure:"POOL_SIZE"`
	ExecTimeOut   time.Duration `mapstructure:"EXEC_TIMEOUT"`
	RelativePath  string        `mapstructure:"RELATIVE_PATH"`
	RootDir       string        `mapstructure:"ROOT_DIR"`
	DBUrl         string        `mapstructure:"DB_URL"`
	MQUrl         string        `mapstructure:"MQ_URL"`
	SelfUpdateURL string        `mapstructure:"SELF_URL"`
	LogOutput     string        `mapstructure:"LOG_OUTPUT"`
	LogLevel      string        `mapstructure:"LOG_LEVEL"`
	DataCenterID  int64         `mapstructure:"DATA_CENTER_ID"`
	NodeName      string        `mapstructure:"NODE_NAME"`
	NodeID        int64         `mapstructure:"NODE_ID"`
}

func Init() error {
	if err := viper.Unmarshal(App); err != nil {
		return err
	}
	defer func() {
		val := reflect.TypeOf(App).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			tag := field.Tag.Get("mapstructure")
			err := os.Unsetenv(tag)
			if err != nil {
				logx.Warnln(err)
			}
		}
	}()
	return App.init()
}

func (c *SConfig) init() error {
	var logfile string
	if c.LogOutput == "file" {
		logfile = filepath.Join(c.LogDir(), utils.ServiceName+".log")
	}
	logx.SetupConsoleLogger(logfile, zap.AddStacktrace(zapcore.FatalLevel))
	level, err := zapcore.ParseLevel(c.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}
	logx.SetLevel(level)

	before, _, found := strings.Cut(c.DBUrl, "://")
	if !found {
		return fmt.Errorf("invalid database url")
	}
	if before == storage.TypeSqlite {
		dir := filepath.Join(c.RootDir, "data")
		if err = utils.EnsureDirExist(dir); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %v", dir, err)
		}
		file := filepath.Join(dir, fmt.Sprintf("%s.db3", utils.ServiceName))
		logx.Infof("%s file: %s", "data", file)
		c.DBUrl = fmt.Sprintf("%s://%s", storage.TypeSqlite, file)
	}

	var dirs = map[string]string{
		"root":      c.RootDir,
		"script":    c.ScriptDir(),
		"log":       c.LogDir(),
		"workspace": c.WorkSpace(),
	}
	for name, dir := range dirs {
		if name == "log" && c.LogOutput != "file" {
			continue
		}
		if err = utils.EnsureDirExist(dir); err != nil {
			return fmt.Errorf("failed to ensure directory %s: %v", dir, err)
		}
		logx.Infof("%s dir: %s", name, dir)
	}
	return nil
}

func (c *SConfig) ScriptDir() string {
	return filepath.Join(c.RootDir, "scripts")
}

func (c *SConfig) LogDir() string {
	return filepath.Join(c.RootDir, "logs")
}

func (c *SConfig) WorkSpace() string {
	return filepath.Join(c.RootDir, "workspace")
}
