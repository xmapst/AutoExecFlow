package server

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xmapst/logx"

	"github.com/xmapst/AutoExecFlow/internal/config"
	"github.com/xmapst/AutoExecFlow/internal/server"
	"github.com/xmapst/AutoExecFlow/internal/utils"
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
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Init(); err != nil {
				return err
			}
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

	viper.AutomaticEnv()
	cmd.Flags().Int64("node_id", 1, "node id")
	cmd.Flags().Int64("data_center_id", 1, "data center id")
	cmd.Flags().String("root_dir", utils.DefaultDir, "root directory")
	cmd.Flags().String("relative_path", "/", "web relative path")
	cmd.Flags().String("node_name", "AutoExecFlow01", "node name")
	cmd.Flags().String("log_output", "file", "log output [file,stdout]")
	cmd.Flags().String("addr", "tcp://0.0.0.0:2376", "listening address.")
	cmd.Flags().String("log_level", "debug", "log level [debug,info,warn,error]")
	cmd.Flags().String("db_url", "sqlite://localhost", "database type. [sqlite,mysql,postgres,sqlserver]")
	cmd.Flags().Duration("exec_timeout", 24*time.Hour, "set the task exec command expire time")
	cmd.Flags().Int("pool_size", runtime.NumCPU()*2, "set the size of the execution work pool.")
	cmd.Flags().String("mq_url", "inmemory://localhost", "message queue url. [inmemory,amqp]")
	cmd.Flags().String("redis_url", "", "redis url.")
	cmd.Flags().String("self_url", "https://oss.yfdou.com/tools/AutoExecFlow", "self Update URL")
	_ = viper.BindPFlags(cmd.Flags())

	return cmd
}
