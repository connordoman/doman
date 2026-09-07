package cmd

import (
	"doman.sh/doman/cmd/git"
	"github.com/spf13/cobra"
)

var GitCommand = &cobra.Command{
	Use:   "git",
	Short: "Git commands",
	Long:  "long description",
	RunE:  runGitCommand,
}

func init() {
	GitCommand.AddCommand(
		git.AuthorCommand,
		git.RemotesCommand,
		git.BranchStatsCommand,
	)

}

func runGitCommand(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
