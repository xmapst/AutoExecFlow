package engine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kardianos/service"
	"github.com/pires/go-proxyproto"
	"github.com/reiver/go-telnet"
	"github.com/soheilhy/cmux"
	"github.com/xmapst/osreapi/internal/cache"
	"github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/engine/worker"
	"github.com/xmapst/osreapi/internal/handlers"
	"github.com/xmapst/osreapi/internal/logx"
	"github.com/xmapst/osreapi/internal/utils"
)

type Program struct {
	listener net.Listener
	cmux     cmux.CMux
	http     *http.Server
	wg       *sync.WaitGroup
}

func New() *Program {
	return &Program{
		wg: new(sync.WaitGroup),
	}
}

func (p *Program) init() error {
	if err := config.App.Load(); err != nil {
		logx.Errorln(err)
		return err
	}

	if !config.App.Debug {
		logx.SetupLogger(filepath.Join(config.App.LogDir, config.App.ServiceName+".log"))
	}

	if config.App.KeyExpire < config.App.ExecTimeOut {
		return fmt.Errorf("the expiration time cannot be less than the execution timeout time")
	}
	if config.App.KeyExpire == config.App.ExecTimeOut || (config.App.KeyExpire/config.App.ExecTimeOut) < 2 {
		config.App.KeyExpire = config.App.ExecTimeOut * 2
	}

	// 调整工作池的大小
	if config.App.PoolSize > worker.DefaultSize {
		worker.SetSize(config.App.PoolSize)
	}
	logx.Infoln("number of workers", worker.GetSize())

	// clear old script
	utils.ClearOldScript(config.App.ScriptDir)

	// 创建临时内存数据库
	if err := cache.New(config.App.DataDir); err != nil {
		logx.Fatalln(err)
		return err
	}

	// 加载自更新数据
	loadSelfUpdateData()

	gin.SetMode(gin.ReleaseMode)
	if config.App.Debug {
		gin.SetMode(gin.DebugMode)
	}
	gin.DisableConsoleColor()
	return nil
}

func (p *Program) Start(service.Service) error {
	err := p.init()
	if err != nil {
		logx.Errorln(err)
		return err
	}
	logx.Infof("listen address %s", config.App.ListenAddress)
	listener, err := net.Listen("tcp", config.App.ListenAddress)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	p.listener = &proxyproto.Listener{Listener: listener}
	p.cmux = cmux.New(p.listener)
	httpL := p.cmux.Match(cmux.HTTP1Fast())
	tcpL := p.cmux.Match(cmux.Any())
	// listen http server
	go p.listenHttpServer(httpL)
	// listen tcp server
	go p.tcpServer(tcpL)
	go func() {
		if err := p.cmux.Serve(); err != nil {
			logx.Errorln(err)
		}
	}()
	return nil
}

func (p *Program) listenHttpServer(listener net.Listener) {
	p.http = &http.Server{
		WriteTimeout: config.App.WebTimeout,
		ReadTimeout:  config.App.WebTimeout,
		IdleTimeout:  config.App.WebTimeout,
		Handler:      handlers.Router(),
	}
	if err := p.http.Serve(listener); err != nil && err != http.ErrServerClosed {
		logx.Fatalln(err)
	}
}

func (p *Program) Stop(service.Service) error {
	logx.Info("shutdown server")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	// close http server
	if p.http != nil {
		_ = p.http.Shutdown(ctx)
	}

	// close cmux server
	if p.cmux != nil {
		p.cmux.Close()
	}

	// close net listener
	if p.listener != nil {
		_ = p.listener.Close()
	}
	p.wg.Wait()

	cache.Close()
	logx.CloseLogger()
	return nil
}

func (p *Program) tcpServer(listener net.Listener) {
	server := telnet.Server{
		Handler: telnet.EchoHandler,
	}
	err := server.Serve(listener)
	if err != nil {
		logx.Errorln(err)
	}
}
