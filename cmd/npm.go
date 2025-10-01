package cmd

import (
	"github.com/connordoman/doman/cmd/npm"
	"github.com/spf13/cobra"
)

var NPMCommand = &cobra.Command{
	Use:   "npm",
	Short: "NPM commands",
	Long:  "NPM commands",
	RunE:  runNPMCommand,
}

func init() {
	NPMCommand.AddCommand(npm.LockfileCommand)
}

func runNPMCommand(cmd *cobra.Command, args []string) error {
	return nil
}
