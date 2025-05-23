package cmd

import (
	"github.com/connordoman/doman/cmd/git"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "doman",
	Short: "A simple CLI application",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(git.AuthorCommand)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
