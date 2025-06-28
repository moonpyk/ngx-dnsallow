package cmd

import (
	"github.com/spf13/viper"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use: "ngxdnsallow",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			slog.SetLogLoggerLevel(slog.LevelDebug)
			if err := rootConfig.ReadInConfig(); err != nil {
				slog.Error(err.Error())
				os.Exit(1)
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
	rootConfig.SetConfigType("yaml")
	rootConfig.SetConfigName("ngxdnsallow")
	rootConfig.AddConfigPath(".")
}
