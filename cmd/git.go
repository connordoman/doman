package cmd

import (
	"github.com/connordoman/doman/cmd/git"
	"github.com/spf13/cobra"
)

var GitCommand = &cobra.Command{
	Use:   "git",
	Short: "Git commands",
	Long:  "long description",
	RunE:  runGitCommand,
}

func init() {
	GitCommand.AddCommand(git.AuthorCommand)
	GitCommand.AddCommand(git.RemotesCommand)
}

func runGitCommand(cmd *cobra.Command, args []string) error {
	return nil
}
