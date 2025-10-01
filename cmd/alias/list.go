package alias

import (
	"fmt"

	"github.com/connordoman/doman/internal/pkg/alias"
	"github.com/spf13/cobra"
)

var AliasListCommand = &cobra.Command{
	Use:   "list",
	Short: "List all aliases currently added via doman",
	RunE:  runAliasListCommand,
}

func init() {

}

func runAliasListCommand(cmd *cobra.Command, args []string) error {
	aliases, err := alias.ListAliases()
	if err != nil {
		return err
	}
	for _, a := range aliases {
		fmt.Println(a)
	}
	return nil
}
