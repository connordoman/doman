package alias

import (
	"fmt"

	"github.com/connordoman/doman/internal/pkg/alias"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var AliasDeleteCommand = &cobra.Command{
	Use:   "delete <alias>",
	Short: "Delete an alias",
	Args:  cobra.ExactArgs(1),
	RunE:  runAliasDeleteCommand,
}

func init() {

}

func runAliasDeleteCommand(cmd *cobra.Command, args []string) error {

	aliasName := args[0]
	if err := alias.DeleteAlias(aliasName); err != nil {
		return err
	}

	fmt.Println(txt.Successf("deleted alias %q", aliasName))
	alias.PrintReloadWarning()
	return nil
}
