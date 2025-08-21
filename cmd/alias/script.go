package alias

import (
	"github.com/connordoman/doman/internal/pkg/alias"
	"github.com/spf13/cobra"
)

var AliasSetupCommand = &cobra.Command{
	Use:   "setup",
	Short: "Refresh the alias loader script",
	RunE:  runAliasSetupCommand,
}

func init() {

}

func runAliasSetupCommand(cmd *cobra.Command, args []string) error {
	if err := alias.Setup(); err != nil {
		return err
	}
	return nil
}
