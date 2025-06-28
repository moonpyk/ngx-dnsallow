package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"moonpyk.net/ngxdnsallow/pkg"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:     "ngxdnsallow",
		Version: pkg.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			levelString := cmd.Flag("log.level").Value.String()
			level := slog.LevelInfo

			if len(levelString) > 0 {
				if err := level.UnmarshalText([]byte(levelString)); err != nil {
					slog.Error("invalid log-level", "level", levelString)
					os.Exit(pkg.ExitInvalidConfiguration)
				}
			}

			slog.SetLogLoggerLevel(level)

			if err := rootConfig.ReadInConfig(); err != nil {
				slog.Error(err.Error())
				os.Exit(pkg.ExitInvalidConfiguration)
			}
		},
	}
	// rootConfig is the program configuration
	rootConfig = viper.New()
	rootLogger = slog.Default()
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	commitString := commit()
	if len(commitString) > 0 {
		rootCmd.Version = fmt.Sprintf("%s [git: %s]", rootCmd.Version, commitString)
	}

	rootCmd.PersistentFlags().String("log.level", "info", "Minimum log level")

	rootConfig.SetConfigType("yaml")
	rootConfig.SetConfigName("ngxdnsallow")
	rootConfig.AddConfigPath(".")
	rootConfig.AddConfigPath("/etc/")
}

func commit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}
