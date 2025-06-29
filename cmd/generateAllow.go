package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"moonpyk.net/ngx-dnsallow/pkg"
	"moonpyk.net/ngx-dnsallow/pkg/config"
)

var (
	// generateAllowCmd represents the generateAllow command
	generateAllowCmd = &cobra.Command{
		Use:     "generate-allow [destination]",
		Aliases: []string{"allow", "gen-allow", "gena"},
		Short:   "Generates allow/deny rules based on configuration",
		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(run(cmd, args))
		},
		Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
	}
	nginxReload     bool
	force           bool
	continueOnError bool
)

func run(cmd *cobra.Command, args []string) int {
	if len(args) == 0 {
		args = []string{"-"}
	}
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
		return pkg.ExitInvalidConfiguration
	}

	if len(cfg.Hosts) == 0 {
		slog.Error("no hosts configured")
		return pkg.ExitInvalidConfiguration
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"# This file is managed by %s v%s, manual changes will be overwritten",
		cmd.Root().Name(),
		cmd.Root().Version,
	))

	sb.WriteString("\n\n")

	for ix, host := range cfg.Hosts {
		if len(host.Dns) == 0 {
			slog.Warn("Dns field is empty, skipping", "index", ix)
			if continueOnError {
				continue
			} else {
				os.Exit(pkg.ExitGenerationError)
			}
		}

		resolved, err := host.LookupIP()
		if err != nil {
			slog.Warn(
				"DNS resolution failed for",
				"host",
				host.Dns,
				"err",
				err,
			)
			if continueOnError {
				continue
			} else {
				os.Exit(pkg.ExitGenerationError)
			}
		}

		s, err := host.RenderAllowLine(resolved)
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
		sb.WriteString(s)
	}

	var dest *os.File

	if args[0] == "-" {
		dest = os.Stdout
	} else {
		var err error
		dest, err = os.OpenFile(
			args[0],
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0o644,
		)
		if err != nil {
			slog.Error("while opening", "file", args[0], "error", err)
			return pkg.ExitGenerationError
		}
		defer func() {
			_ = dest.Close()
		}()
	}

	_, err := dest.Write([]byte(sb.String()))
	if err != nil {
		slog.Error("while writing to", "file", args[0], "error", err)
		return pkg.ExitGenerationError
	}

	return pkg.ExitSuccess
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
