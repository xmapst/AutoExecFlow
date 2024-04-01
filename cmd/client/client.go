package client

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/xmapst/osreapi/pkg/logx"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "client",
		Short: "a self-sufficient executor",
		Aliases: []string{
			"cli",
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logx.SetupLogger("", zap.AddStacktrace(zap.ErrorLevel))
			logx.Infoln("under development, please stay tuned")
			return nil
		},
	}
	return cmd
}
