package cmd

import (
	"github.com/spf13/viper"
	"log/slog"
	"moonpyk.net/ngxdnsallow/pkg"
	"os"

	"github.com/spf13/cobra"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use: "ngxdnsallow",
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
	rootCmd.PersistentFlags().String("log.level", "info", "Minimum log level")

	rootConfig.SetConfigType("yaml")
	rootConfig.SetConfigName("ngxdnsallow")
	rootConfig.AddConfigPath(".")
	rootConfig.AddConfigPath("/etc/")
}
