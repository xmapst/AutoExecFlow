//go:build (windows || linux || darwin) && (amd64 || arm64 || 386) && cgo

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/xmapst/AutoExecFlow/cmd/server"
	"github.com/xmapst/AutoExecFlow/pkg/info"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

const longText = `An API for cross-platform custom orchestration of execution steps without any third-party dependencies. 
Based on DAG, it implements the scheduling function of sequential execution of dependent steps and concurrent execution of non-dependent steps.`

func main() {
	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Operating system remote execution interface",
		Long:  longText,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		Version: info.Version,
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.SetHelpTemplate(`{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)
	cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	_ = cmd.PersistentFlags().MarkShorthandDeprecated("help", "please use --help")
	cmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	cmd.AddCommand(
		server.New(),
		&cobra.Command{
			Use:   "version",
			Short: "print version information and quit",
			RunE: func(cmd *cobra.Command, args []string) error {
				info.PrintHeadInfo()
				return nil
			},
		})

	if err := cmd.Execute(); err != nil {
		logx.Fatalln(err)
	}
}
