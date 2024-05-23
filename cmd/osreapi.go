package main

import (
<<<<<<< HEAD
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin/v2"
	"github.com/kardianos/service"

	info "github.com/xmapst/osreapi"
	"github.com/xmapst/osreapi/internal/config"
	_ "github.com/xmapst/osreapi/internal/config"
	"github.com/xmapst/osreapi/internal/engine"
)

func init() {
	// flags
	kingpin.Flag(
		"addr",
		"host:port for execution.",
	).Default(":2376").StringVar(&config.App.ListenAddress)
	kingpin.Flag(
		"normal",
		"Normal wait for all task execution to complete",
	).Default("false").BoolVar(&config.App.Normal)
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
}

// @title           OSReApi
// @version         1.0
// @description     This is a OS Remote Executor Api Server.

// @contact.name   osreapi
// @contact.url    https://github.com/xmapst/osreapi/issues

// @license.name  GPL-3.0
// @license.url   https://github.com/xmapst/osreapi/blob/main/LICENSE
func main() {
	kingpin.Version(info.VersionInfo())
	kingpin.HelpFlag.Short('h')
	kingpin.Command("run", "Run server").Action(run)
	kingpin.Parse()
}

func run(*kingpin.ParseContext) (err error) {
	info.PrintHeadInfo()
	var svc service.Service
	svc, err = service.New(engine.New(), &service.Config{
		Name:        config.App.ServiceName,
		DisplayName: "OSReApi",
		Description: "OS Remote Executor Api",
	})
	if err != nil {
		return err
	}
	err = svc.Run()
	if err != nil {
		return err
	}
	return nil
=======
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/xmapst/osreapi/cmd/client"
	"github.com/xmapst/osreapi/cmd/server"
	"github.com/xmapst/osreapi/pkg/info"
)

const longText = `The remote executor (OSReApi) provides API remote operation mode,batch execution of Shell, Powershell, Python and other commands, and easily completes common management tasks such as running automated operation and maintenance scripts, polling processes, installing or uninstalling software, updating applications, and installing patches.`

func main() {
	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Operating system remote execution interface",
		Long:  longText,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		Version: info.Version,
	}

	cmd.SetFlagErrorFunc(flagErrorFunc)
	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.SetHelpTemplate(`{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)
	cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	_ = cmd.PersistentFlags().MarkShorthandDeprecated("help", "please use --help")
	cmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	cmd.AddCommand(server.New())
	cmd.AddCommand(client.New())
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "print version information and quit",
		RunE: func(cmd *cobra.Command, args []string) error {
			info.PrintHeadInfo()
			return nil
		},
	})

	if err := cmd.Execute(); err != nil {
		os.Exit(128)
	}
}

func flagErrorFunc(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}

	usage := ""
	if cmd.HasSubCommands() {
		usage = "\n\n" + cmd.UsageString()
	}
	return fmt.Errorf("%s\nSee '%s --help'.%s", err, cmd.CommandPath(), usage)
>>>>>>> githubB
}
