package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/udhos/equalfile"

	"github.com/spf13/cobra"

	"moonpyk.net/ngx-dnsallow/pkg"
	"moonpyk.net/ngx-dnsallow/pkg/model"
)

var (
	// generateAllowCmd represents the generate-allow command
	generateAllowCmd = &cobra.Command{
		Use:     "generate-allow [destination]",
		Aliases: []string{"allow", "gen-allow", "gena"},
		Short:   "Generates allow/deny rules based on configuration",
		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(run(cmd, args))
		},
		Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
	}
	flagNginxReload     bool // flagNginxReload indicates whether to reload nginx configuration after generation
	flagForce           bool // flagForce indicates whether to force output even if no changes are detected
	flagContinueOnError bool // flagContinueOnError indicates whether to continue processing even if an error occurs
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
		flagNginxReload,
		"force",
		flagForce,
	)

	var cfg model.Config

	if err := rootConfig.Unmarshal(&cfg); err != nil {
		return pkg.ExitInvalidConfiguration
	}

	if len(cfg.Hosts) == 0 {
		rootLogger.Error("no host configured")

		return pkg.ExitInvalidConfiguration
	}

	if flagNginxReload {
		if err := cfg.Nginx.ConfigTest(); err != nil {
			rootLogger.Error(
				"nginx pre-config test returned an error, aborting",
				"error",
				err,
			)

			return pkg.ExitNginxError
		}

		rootLogger.Debug("nginx pre-config test succeeded")
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"# This file is managed by %s v%s, manual changes will be overwritten",
		cmd.Root().Name(),
		cmd.Root().Version,
	))

	sb.WriteString("\n\n")

	for ix, host := range cfg.Hosts {
		if len(host.Hostname) == 0 {
			rootLogger.Warn("Hostname field is empty, skipping", "index", ix)

			if flagContinueOnError {
				continue
			} else {
				return pkg.ExitGenerationError
			}
		}

		resolved, err := host.LookupIP()
		if err != nil {
			rootLogger.Warn(
				"DNS resolution failed for",
				"host",
				host.Hostname,
				"err",
				err,
			)

			if flagContinueOnError {
				continue
			} else {
				return pkg.ExitGenerationError
			}
		}

		s, err := host.RenderAllowLine(resolved)
		if err != nil {
			rootLogger.Warn(
				"rendering error",
				"index",
				ix,
				"host",
				host.Hostname,
				"type",
				host.EnsureType(),
				"error",
				err.Error(),
			)

			if flagContinueOnError {
				continue
			} else {
				return pkg.ExitGenerationError
			}
		}

		rootLogger.Debug(
			"generated",
			"index",
			ix,
			"host",
			host.Hostname,
			"lines",
			s,
		)
		sb.WriteString(s)
	}

	var dest *os.File

	if args[0] != "-" {
		var err error

		dest, err = os.OpenFile(
			args[0],
			os.O_CREATE|os.O_RDWR,
			0o644,
		)
		if err != nil {
			rootLogger.Error(
				"while opening",
				"file",
				args[0],
				"error",
				err,
			)

			return pkg.ExitGenerationError
		}

		defer func() {
			_ = dest.Close()
		}()
	} else {
		dest = os.Stdout
	}

	result := sb.String()

	if dest != os.Stdout {
		if !flagForce {
			cmp := equalfile.New(nil, equalfile.Options{})

			equal, err := cmp.CompareReader(
				dest,
				strings.NewReader(result),
			)
			if err == nil && equal {
				rootLogger.Debug("destination and current output are equal, no need to continue")

				return pkg.ExitSuccess
			}
		}

		if err := dest.Truncate(0); err != nil {
			rootLogger.Error(
				"error while truncating",
				"file",
				dest.Name(),
				"error",
				err,
			)

			return pkg.ExitGenerationError
		}

		rootLogger.Info(
			"configuration changed or is new, writing to",
			"file",
			dest.Name(),
			"force",
			flagForce,
		)

		// really, we don't await an error here, but as good practice...
		if _, err := dest.Seek(0, io.SeekStart); err != nil {
			rootLogger.Error(
				"error while seeking to the beginning of the",
				"file",
				dest.Name(),
				"error",
				err,
			)

			return pkg.ExitGenerationError
		}
	}

	_, err := dest.Write([]byte(result))
	if err != nil {
		rootLogger.Error("while writing to", "file", args[0], "error", err)

		return pkg.ExitGenerationError
	}

	if flagNginxReload {
		if err = cfg.Nginx.ConfigTest(); err != nil {
			rootLogger.Error(
				"nginx config test returned an error, aborting",
				"error",
				err,
			)

			return pkg.ExitNginxError
		}

		rootLogger.Debug("nginx config test succeeded")

		if err = cfg.Nginx.ConfigReload(); err != nil {
			rootLogger.Error(
				"nginx config reload returned an error, aborting",
				"error",
				err,
			)

			return pkg.ExitNginxError
		}

		rootLogger.Debug("nginx config reload succeeded")
	}

	return pkg.ExitSuccess
}

func init() {
	flags := generateAllowCmd.Flags()

	flags.BoolVarP(
		&flagNginxReload,
		"nginx.reload",
		"r",
		true,
		"Reload nginx configuration if changes are made",
	)
	flags.BoolVarP(
		&flagForce,
		"force",
		"f",
		false,
		"Emit output even if no change is detected",
	)
	flags.BoolVarP(
		&flagContinueOnError,
		"continue",
		"C",
		false,
		"Continue on any generation error",
	)
	rootCmd.AddCommand(generateAllowCmd)
}
