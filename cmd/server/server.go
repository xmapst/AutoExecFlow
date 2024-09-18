package server

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"

	"github.com/xmapst/AutoExecFlow/internal/server"
	"github.com/xmapst/AutoExecFlow/internal/server/config"
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

	cmd.Flags().BoolVarP(&config.App.Debug, "debug", "d", false, "debug mode.")
	cmd.Flags().StringVar(&config.App.DBType, "db_type", "sqlite", "database type. [sqlite]")
	cmd.Flags().DurationVar(&config.App.ExecTimeOut, "exec_timeout", 24*time.Hour, "set the task exec command expire time")
	cmd.Flags().StringVarP(&config.App.ListenAddress, "addr", "a", "tcp://0.0.0.0:2376", "listening address.")
	cmd.Flags().BoolVarP(&config.App.Normal, "normal", "n", false, "wait for all task execution to complete.")
	cmd.Flags().IntVarP(&config.App.PoolSize, "pool_size", "p", runtime.NumCPU()*2, "set the size of the execution work pool.")
	cmd.Flags().StringVarP(&config.App.SelfUpdateURL, "self_url", "s", "https://oss.yfdou.com/tools/AutoExecFlow", "self Update URL")
	cmd.Flags().DurationVarP(&config.App.WebTimeout, "timeout", "t", 120*time.Second, "maximum duration before timing out read/write/idle.")
	cmd.Flags().StringVarP(&config.App.RelativePath, "relative_path", "r", "/", "web relative path")
	cmd.Flags().StringVar(&config.App.RootDir, "root_dir", utils.DefaultDir, "root directory")

	return cmd
}
