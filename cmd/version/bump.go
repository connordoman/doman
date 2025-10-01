package version

import (
	"fmt"

	"github.com/connordoman/doman/internal/config"
	"github.com/spf13/cobra"
)

var BumpCommand = &cobra.Command{
	Use:   "bump [major|minor|patch]",
	Short: "Bump the version number",
	Long:  "Bump the version number",
	RunE:  runBumpCommand,
	Args:  cobra.ExactArgs(1),
}

func init() {

}

const (
	versionFileLocation = "./VERSION"
)

func runBumpCommand(cmd *cobra.Command, args []string) error {
	versionInfo, err := config.OpenVersionFile(versionFileLocation)
	if err != nil {
		return err
	}

	switch args[0] {
	case "major":
		versionInfo.Bump(config.BumpMajor)
	case "minor":
		versionInfo.Bump(config.BumpMinor)
	case "patch":
		versionInfo.Bump(config.BumpPatch)
	default:
		return fmt.Errorf("unknown version segment: %s", args[0])
	}

	return nil
}
