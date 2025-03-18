package cmd

import (
	"github.com/agmanchon/blockchainmonitor/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName = "Blockchain Monitor"
)

var rootCmd = &cobra.Command{
	Use:   "blockchain Monitor",
	Short: "Daemon for " + appName,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	registerFlags()
	registerCommands()
}

func registerFlags() {
	rootCmd.PersistentFlags().String(config.ConfigPath, "", "Configuration path")
	err := viper.BindPFlag(config.ConfigPath, rootCmd.PersistentFlags().Lookup(config.ConfigPath))
	if err != nil {
		panic(err)
	}
}

func registerCommands() {
	rootCmd.AddCommand(startCmd)

}
