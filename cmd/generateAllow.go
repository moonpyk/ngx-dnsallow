package cmd

import (
	"github.com/spf13/cobra"
	"log/slog"
	"moonpyk.net/ngxdnsallow/pkg/config"
	"net"
	"os"
)

var (
	// generateAllowCmd represents the generateAllow command
	generateAllowCmd = &cobra.Command{
		Use:     "generate-allow",
		Aliases: []string{"allow", "gen-allow", "gena"},
		Short:   "A brief description of your command",
		Run:     run,
	}
	nginxReload bool
	force       bool
)

func run(_ *cobra.Command, _ []string) {
	rootLogger.Debug(
		"Using config",
		"file",
		rootConfig.ConfigFileUsed(),
		"nginx.reload",
		nginxReload,
		"force",
		force,
	)

	var cfg config.Config

	if err := rootConfig.Unmarshal(&cfg); err != nil {
		os.Exit(1)
		return
	}

	if len(cfg.Hosts) == 0 {
		slog.Warn("no hosts configured")
		os.Exit(1)
	}

	for ix, host := range cfg.Hosts {
		if len(host.Dns) == 0 {
			slog.Warn("Dns field is empty, skipping", "index", ix)
			continue
		}

		addrs, err := net.LookupIP(host.Dns)

		if err != nil {
			continue
		}

		err, s := host.RenderAllowLine(addrs)

		if err != nil {
			slog.Warn(
				"rendering error",
				"index",
				ix,
				"host",
				host.Dns,
				"type",
				host.EnsureType(),
				"error",
				err.Error(),
			)
			continue
		}

		slog.Debug(
			"generated",
			"index",
			ix,
			"host",
			host.Dns,
			"lines",
			s,
		)
	}
}

func init() {
	generateAllowCmd.Flags().BoolVarP(
		&nginxReload,
		"nginx.reload",
		"r",
		true,
		"Reload nginx configuration if changes are made",
	)
	generateAllowCmd.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Emit output even if no change is detected",
	)
	rootCmd.AddCommand(generateAllowCmd)
}
