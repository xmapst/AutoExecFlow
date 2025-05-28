package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
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
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/queues"
	"github.com/xmapst/AutoExecFlow/internal/server/api"
	"github.com/xmapst/AutoExecFlow/internal/server/tus"
	"github.com/xmapst/AutoExecFlow/internal/storage"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/internal/worker"
	"github.com/xmapst/AutoExecFlow/pkg/listeners"
)

type sProgram struct {
	ctx       context.Context
	cancel    context.CancelFunc
	sHash     []byte
	sURL      string
	listeners []net.Listener
	http      *http.Server
	wg        *sync.WaitGroup
	cron      *cron.Cron
}

func New() service.Interface {
	p := &sProgram{
		sURL: strings.TrimSuffix(config.App.SelfUpdateURL, "/"),
		cron: cron.New(),
		wg:   new(sync.WaitGroup),
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	// 获取当前程序的hash
	p.localSha256sum()
	return p
}

func (p *sProgram) init() error {
	if _, err := p.cron.AddFunc("@every 1m", func() {
		debug.FreeOSMemory()
	}); err != nil {
		logx.Errorln(err)
		return err
	}

	// 启动自更新监控
	p.selfUpdate()

	// setup queue
	if err := queues.New(config.App.NodeName, config.App.MQUrl); err != nil {
		return fmt.Errorf("failed to setup queue: %v", err)
	}

	// 创建临时内存数据库
	if err := storage.New(
		config.App.DataCenterID,
		config.App.NodeID,
		config.App.DBUrl,
	); err != nil {
		logx.Errorln(err)
		return err
	}

	// 修正当前节点重启前的数据
	return storage.FixDatabase(config.App.NodeName)
}

func (p *sProgram) Start(service.Service) error {
	p.cron.Start()
	err := p.init()
	if err != nil {
		return err
	}
	err = p.startWorker()
	if err != nil {
		return err
	}
	return p.startAPI()
}

func (p *sProgram) startWorker() error {
	// 调整工作池的大小
	worker.SetSize(config.App.PoolSize)
	logx.Infoln("number of workers", worker.GetSize())

	// clear old script
	utils.ClearDir(config.App.ScriptDir())

	// clear old workspace
	utils.ClearDir(config.App.WorkSpace())

	// 启动任务执行器
	return worker.Start(p.ctx)
}

func (p *sProgram) startAPI() error {
	// 首次激活事件监听
	ctx, cancel := context.WithCancel(p.ctx)
	defer cancel()
	if err := queues.SubscribeEvent(ctx, func(data string) error {
		logx.Debugln(data)
		return nil
	}); err != nil {
		logx.Errorln(err)
		return err
	}
	return p.startServer()
}

func (p *sProgram) startServer() error {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	handler, err := api.New(config.App.RelativePath)
	if err != nil {
		logx.Errorln(err)
		return err
	}
	var filesPath = path.Join(config.App.RelativePath, "/api/v1/files/")
	if err = tus.Init(config.App.WorkSpace(), filesPath, config.App.RedisUrl); err != nil {
		logx.Errorln(err)
		return err
	}
	handler.Any(strings.TrimSuffix(filesPath, "/"), gin.WrapH(tus.TunServer))
	handler.Any(path.Join(filesPath, "/*any"), gin.WrapH(tus.TunServer))

	p.http = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 120 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    15 << 20, // 15MB
		BaseContext: func(_ net.Listener) context.Context {
			return p.ctx
		},
	}

	_ = retry.Do(
		func() error {
			if err := p.loadListeners([]string{
				config.App.Address,
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

func (p *sProgram) loadListeners(hosts []string) error {
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

func (p *sProgram) close() {
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

func (p *sProgram) Stop(service.Service) error {
	logx.Infoln("stop service")
	p.close()
	p.wg.Wait()
	p.cron.Stop()

	ctx, cancel := context.WithTimeout(p.ctx, time.Second*15)
	defer cancel()
	logx.Infoln("shutdown queue")
	queues.Shutdown(ctx)

	logx.Infoln("shutdown worker")
	worker.Shutdown()

	logx.Infoln("close storage")
	if err := storage.Close(); err != nil {
		logx.Errorln(err)
	}
	p.cancel()
	logx.Infoln("service stopped")
	logx.CloseLogger()
	time.Sleep(300 * time.Millisecond)
	return nil
}
