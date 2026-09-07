package cmd

import (
	"fmt"

	"doman.sh/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var VersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Display the current version of the application",
	RunE:  runVersionCommand,
}

func init() {
}

func runVersionCommand(cmd *cobra.Command, args []string) error {
	fmt.Println(pkg.VersionString())
	return nil
}
