package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"moonpyk.net/ngx-dnsallow/pkg"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootConfig = viper.New()    // rootConfig is the program configuration
	rootLogger = slog.Default() // rootLogger is the program default logger
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:     "ngx-dnsallow",
		Version: pkg.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			levelString := cmd.Flag("log.level").Value.String()
			level := slog.LevelInfo

			if len(levelString) > 0 {
				if err := level.UnmarshalText([]byte(levelString)); err != nil {
					rootLogger.Error("invalid log.level", "level", levelString)
					os.Exit(pkg.ExitInvalidConfiguration)
				}
			}

			slog.SetLogLoggerLevel(level)

			configPath := cmd.Flag("config").Value.String()
			if len(configPath) > 0 {
				rootConfig.SetConfigFile(configPath)
			}

			if err := rootConfig.ReadInConfig(); err != nil {
				rootLogger.Error(err.Error())
				os.Exit(pkg.ExitInvalidConfiguration)
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	err := rootCmd.Execute()
	if err != nil {
		return 1
	}

	return 0
}

func init() {
	rootConfig.SetConfigType("yaml")
	rootConfig.SetConfigName("ngx-dnsallow")
	rootConfig.AddConfigPath(".")
	rootConfig.AddConfigPath("/etc/")

	commitString := gitCommit()
	if len(commitString) > 0 {
		rootCmd.Version = fmt.Sprintf("%s [git: %s]", rootCmd.Version, commitString)
	}

	pflags := rootCmd.PersistentFlags()
	pflags.StringP(
		"config",
		"c",
		"",
		"config file path override",
	)
	pflags.String(
		"log.level",
		"info",
		"Minimum log level",
	)
}

// gitCommit returns the current git commit hash if available
func gitCommit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}
