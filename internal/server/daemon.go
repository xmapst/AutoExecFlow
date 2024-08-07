package server

import (
	"context"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/pires/go-proxyproto"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"github.com/xmapst/AutoExecFlow/internal/api"
	"github.com/xmapst/AutoExecFlow/internal/server/config"
	"github.com/xmapst/AutoExecFlow/internal/server/listeners"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type Program struct {
	sHash     []byte
	sURL      string
	listeners []net.Listener
	http      *http.Server
	wg        *sync.WaitGroup
	cron      *cron.Cron
}

func New() *Program {
	p := &Program{
		sURL: strings.TrimSuffix(config.App.SelfUpdateURL, "/"),
		cron: cron.New(),
		wg:   new(sync.WaitGroup),
	}
	// 获取当前程序的hash
	p.localSha256sum()
	return p
}

func (p *Program) init() error {
	if err := config.App.Init(); err != nil {
		logx.Errorln(err)
		return err
	}

	if _, err := p.cron.AddFunc("@every 1m", func() {
		debug.FreeOSMemory()
	}); err != nil {
		logx.Errorln(err)
		return err
	}

	// 启动自更新监控
	p.selfUpdate()

	// 调整工作池的大小
	if config.App.PoolSize > worker.DefaultSize {
		worker.SetSize(config.App.PoolSize)
	}
	logx.Infoln("number of workers", worker.GetSize())

	// clear old script
	utils.ClearDir(config.App.ScriptDir())

	// clear old workspace
	utils.ClearDir(config.App.WorkSpace())

	// 创建临时内存数据库
	if err := storage.New(config.App.DBType, config.App.DataDir()); err != nil {
		logx.Fatalln(err)
		return err
	}

	return nil
}

func (p *Program) Start(service.Service) error {
	p.cron.Start()
	err := p.init()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	return p.startServer()
}

func (p *Program) startServer() error {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	if config.App.Debug {
		gin.SetMode(gin.DebugMode)
	}
	p.http = &http.Server{
		WriteTimeout: config.App.WebTimeout,
		ReadTimeout:  config.App.WebTimeout,
		IdleTimeout:  config.App.WebTimeout,
		Handler:      router.New(config.App.RelativePath),
	}

	_ = retry.Do(
		func() error {
			if err := p.loadListeners([]string{
				config.App.ListenAddress,
				utils.PipeName,
			}); err != nil {
				logx.Errorln(err)
				return err
			}
			return nil
		},
		retry.Attempts(0),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			_max := time.Duration(n)
			if _max > 8 {
				_max = 8
			}
			duration := time.Second * _max * _max
			return duration
		}),
	)
	for _, ln := range p.listeners {
		p.wg.Add(1)
		go func(ln net.Listener) {
			defer p.wg.Done()
			if err := p.http.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logx.Errorln(err)
			}
		}(ln)
	}

	return nil
}

func (p *Program) loadListeners(hosts []string) error {
	for _, ln := range p.listeners {
		_ = ln.Close()
	}
	p.listeners = []net.Listener{}
	for _, host := range hosts {
		proto, addr, ok := strings.Cut(host, "://")
		if !ok {
			logx.Warnf("bad format %s, expected PROTO://ADDR", host)
			proto = "tcp"
			addr = host
		}
		ln, err := listeners.Init(proto, addr, nil)
		if err != nil {
			logx.Errorln(err)
			return err
		}
		logx.Infof("Listener created on %s (%s)", proto, addr)
		p.listeners = append(p.listeners, &proxyproto.Listener{Listener: ln})
	}
	return nil
}

func (p *Program) close() {
	logx.Info("shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// close http server
	if p.http != nil {
		_ = p.http.Shutdown(ctx)
	}

	// close net listener
	if p.listeners != nil {
		for _, ln := range p.listeners {
			_ = ln.Close()
		}
	}
}

func (p *Program) Stop(service.Service) error {
	logx.Infoln("receive a shutdown event and start shutting down")
	p.close()
	p.wg.Wait()
	if config.App.Normal {
		logx.Infoln("wait for workers to converge")
		worker.StopWait()
	}
	p.cron.Stop()
	logx.Infoln("put data to disk and close data storage")
	if err := storage.Close(); err != nil {
		logx.Errorln(err)
	}
	logx.CloseLogger()
	time.Sleep(1 * time.Second)
	return nil
}
