package cmd

import (
	"github.com/ysicing/spot/cmd/flags"

	"github.com/ergoapi/util/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	globalFlags *flags.GlobalFlags
)

func init() {
	cobra.OnInitialize(initConfig)
}

func Execute() {
	rootCmd := BuildRoot()
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

// BuildRoot creates a new root command from the
func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()
	persistentFlags := rootCmd.PersistentFlags()
	globalFlags = flags.SetGlobalFlags(persistentFlags)

	// Add main commands
	rootCmd.AddCommand(cmdNew())
	rootCmd.AddCommand(cmdList())
	rootCmd.AddCommand(cmdDestroy())
	rootCmd.AddCommand(cmdRestart())
	rootCmd.AddCommand(cmdScan())
	rootCmd.AddCommand(cmdImage())
	return rootCmd
}

// NewRootCmd returns a new root command
func NewRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "spot",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "腾讯云虚拟机管理工具",
		Version:       "0.0.4",
		PersistentPreRunE: func(cobraCmd *cobra.Command, args []string) error {
			if globalFlags.Debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
	}
}

func initConfig() {
	viper.SetConfigFile(globalFlags.ConfigPath)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		logrus.Infof("Using config file: %v", color.SGreen(viper.ConfigFileUsed()))
	}
}
