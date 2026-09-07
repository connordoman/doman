package cmd

import (
	"doman.sh/cmd/config"
	"github.com/spf13/cobra"
)

var ConfigCommand = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  "This command allows you to manage configuration settings for the application.",
}

func init() {
	ConfigCommand.AddCommand(
		config.FindConfigCommand,
		config.EditConfigCommand,
		config.InitConfigCommand,
	)
}
