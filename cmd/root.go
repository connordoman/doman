package cmd

import (
	"github.com/connordoman/doman/cmd/completions"
	"github.com/connordoman/doman/cmd/git"
	"github.com/connordoman/doman/cmd/npm"
	"github.com/connordoman/doman/cmd/sys"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "doman",
	Short: "A simple CLI application for domain-management tasks",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		pkg.SetEcho(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Set up
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Flags
	rootCmd.PersistentFlags().BoolP("echo", "e", false, "Print the underlying commands being executed")

	// Commands
	rootCmd.AddCommand(completions.CompletionsCommand)
	rootCmd.AddCommand(git.AuthorCommand)
	rootCmd.AddCommand(git.RemotesCommand)
	rootCmd.AddCommand(npm.LockfileCommand)
	rootCmd.AddCommand(sys.IPCommand)
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
