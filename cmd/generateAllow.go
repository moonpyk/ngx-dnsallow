package cmd

import (
	"github.com/spf13/cobra"
	"log/slog"
	"moonpyk.net/ngxdnsallow/pkg"
	"moonpyk.net/ngxdnsallow/pkg/config"
	"net"
	"os"
)

var (
	// generateAllowCmd represents the generateAllow command
	generateAllowCmd = &cobra.Command{
		Use:     "generate-allow",
		Aliases: []string{"allow", "gen-allow", "gena"},
		Short:   "Generates allow/deny rules based on configuration",
		Run:     run,
	}
	nginxReload     bool
	force           bool
	continueOnError bool
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
		slog.Error("no hosts configured")
		os.Exit(pkg.ExitInvalidConfiguration)
	}

	for ix, host := range cfg.Hosts {
		if len(host.Dns) == 0 {
			slog.Warn("Dns field is empty, skipping", "index", ix)
			if continueOnError {
				continue
			} else {
				os.Exit(pkg.ExitGenerationError)
			}
		}

		addrs, err := net.LookupIP(host.Dns)

		if err != nil {
			slog.Warn("DNS resolution failed for", "host", host.Dns, "err", err)
			if continueOnError {
				continue
			} else {
				os.Exit(pkg.ExitGenerationError)
			}
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

			if continueOnError {
				continue
			} else {
				os.Exit(pkg.ExitGenerationError)
			}
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
	flags := generateAllowCmd.Flags()

	flags.BoolVarP(
		&nginxReload,
		"nginx.reload",
		"r",
		true,
		"Reload nginx configuration if changes are made",
	)
	flags.BoolVarP(
		&force,
		"force",
		"f",
		false,
		"Emit output even if no change is detected",
	)
	flags.BoolVarP(
		&continueOnError,
		"continue",
		"C",
		false,
		"Continue on any generation error",
	)
	rootCmd.AddCommand(generateAllowCmd)
}
