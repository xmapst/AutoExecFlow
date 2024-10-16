package server

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/server"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "server",
		Aliases: []string{
			"run",
		},
		Short: "start server",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := filepath.Abs(os.Args[0])
			if err != nil {
				logx.Errorln(err)
				return err
			}
			svc, err := service.New(server.New(), &service.Config{
				Name:        utils.ServiceName,
				DisplayName: utils.ServiceName,
				Description: "Operating System Remote Executor Api",
				Executable:  name,
				Arguments:   os.Args[1:],
			})
			if err != nil {
				return err
			}
			err = svc.Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config.App.RootDir, "root_dir", utils.DefaultDir, "root directory")
	cmd.Flags().StringVar(&config.App.RelativePath, "relative_path", "/", "web relative path")
	cmd.Flags().StringVar(&config.App.LogOutput, "log_output", "file", "log output [file,stdout]")
	cmd.Flags().StringVar(&config.App.Address, "addr", "tcp://0.0.0.0:2376", "listening address.")
	cmd.Flags().StringVar(&config.App.LogLevel, "log_level", "debug", "log level [debug,info,warn,error]")
	cmd.Flags().StringVar(&config.App.DBUrl, "db_url", "sqlite://localhost", "database type. [sqlite,mysql]")
	cmd.Flags().IntVar(&config.App.PoolSize, "pool_size", runtime.NumCPU()*2, "set the size of the execution work pool.")
	cmd.Flags().StringVar(&config.App.MQUrl, "mq_url", "inmemory://localhost", "message queue url. [inmemory,amqp]")
	cmd.Flags().DurationVar(&config.App.ExecTimeOut, "exec_timeout", 24*time.Hour, "set the task exec command expire time")
	cmd.Flags().StringVar(&config.App.SelfUpdateURL, "self_url", "https://oss.yfdou.com/tools/AutoExecFlow", "self Update URL")

	return cmd
}
