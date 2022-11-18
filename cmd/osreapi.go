package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
	info "github.com/xmapst/osreapi"
	"github.com/xmapst/osreapi/cache"
	"github.com/xmapst/osreapi/config"
	_ "github.com/xmapst/osreapi/config"
	"github.com/xmapst/osreapi/engine"
	"github.com/xmapst/osreapi/handlers"
	"github.com/xmapst/osreapi/utils"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func init() {
	// flags
	kingpin.Flag(
		"addr",
		"host:port for execution.",
	).Default(":2376").StringVar(&config.App.ListenAddress)
	kingpin.Flag(
		"debug",
		"Enable debug messages",
	).Default("false").BoolVar(&config.App.Debug)
	kingpin.Flag(
		"root",
		"Working root directory",
	).Default(filepath.Join(os.TempDir(), config.App.ServiceName)).StringVar(&config.App.RootDir)
	kingpin.Flag(
		"key_expire",
		`Set the database key expire time. Example: "key_expire=1h"`,
	).Default("48h").DurationVar(&config.App.KeyExpire)
	kingpin.Flag(
		"exec_timeout",
		`Set the exec command expire time. Example: "exec_timeout=30m"`,
	).Default("24h").DurationVar(&config.App.ExecTimeOut)
	kingpin.Flag(
		"timeout",
		"Timeout for calling endpoints on the engine",
	).Default("30s").DurationVar(&config.App.WebTimeout)
	kingpin.Flag(
		"max-requests",
		"Maximum number of concurrent requests. 0 to disable.",
	).Default("0").Int64Var(&config.App.MaxRequests)
	kingpin.Flag(
		"pool_size",
		"Set the size of the execution work pool.",
	).Default("30").IntVar(&config.App.PoolSize)
	// log format init
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			file = fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
			_f := strings.Split(frame.Function, ".")
			function = _f[len(_f)-1]
			return
		},
	})
}

type program struct {
	svr *http.Server
}

func (p *program) init() error {
	info.PrintHeadInfo()
	logrus.SetOutput(os.Stdout)
	if !config.App.Debug {
		logrus.SetOutput(utils.LogOutput(config.App.LogDir, config.App.ServiceName))
	}

	if config.App.KeyExpire < config.App.ExecTimeOut {
		return fmt.Errorf("the expiration time cannot be less than the execution timeout time")
	}
	if config.App.KeyExpire == config.App.ExecTimeOut || (config.App.KeyExpire/config.App.ExecTimeOut) < 2 {
		config.App.KeyExpire = config.App.ExecTimeOut * 2
	}
	// clear old script
	utils.ClearOldScript(config.App.ScriptDir)
	// 创建临时内存数据库
	cache.Init()
	// 创建池
	engine.NewExecPool(config.App.PoolSize)
	// 加载自更新数据
	engine.LoadSelfUpdateData()
	gin.SetMode(gin.ReleaseMode)
	if config.App.Debug {
		gin.SetMode(gin.DebugMode)
	}
	gin.DisableConsoleColor()
	return nil
}

func (p *program) Start(service.Service) error {
	err := p.init()
	if err != nil {
		logrus.Errorln(err)
		return err
	}
	p.svr = &http.Server{
		Addr:         config.App.ListenAddress,
		WriteTimeout: config.App.WebTimeout,
		ReadTimeout:  config.App.WebTimeout,
		IdleTimeout:  config.App.WebTimeout,
		Handler:      handlers.Router(),
	}
	go p.listenAndServe()
	return nil
}

func (p *program) listenAndServe() {
	logrus.Infof("listen address %s", p.svr.Addr)
	if err := p.svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalln(err)
	}
}

func (p *program) Stop(service.Service) error {
	logrus.Info("shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_ = p.svr.Shutdown(ctx)
	cache.Close()
	return nil
}

func main() {
	kingpin.Version(info.VersionInfo())
	kingpin.HelpFlag.Short('h')
	kingpin.Command("run", "Run server").Action(run)
	kingpin.Parse()
}

func run(*kingpin.ParseContext) error {
	err := config.App.Load()
	if err != nil {
		return err
	}
	var svc service.Service
	svc, err = service.New(&program{}, &service.Config{
		Name:        config.App.ServiceName,
		DisplayName: "Q1Autoops System Remote Executor",
		Description: "Q1Autoops System Remote Executor",
	})
	if err != nil {
		return err
	}
	err = svc.Run()
	if err != nil {
		return err
	}
	return nil
}
