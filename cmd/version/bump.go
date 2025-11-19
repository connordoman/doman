package version

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var BumpCommand = &cobra.Command{
	Use:   "bump [major|minor|patch]",
	Short: "Bump the version number",
	Long:  "Bump the version number",
	RunE:  runBumpCommand,
	Args:  cobra.MaximumNArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		verboseFlag, _ := cmd.Flags().GetBool("verbose")
		if verboseFlag {
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(io.Discard)
		}
		return nil
	},
}

func init() {
	BumpCommand.Flags().BoolP("dry-run", "d", false, "Dry run the bump")
	BumpCommand.Flags().StringP("file", "f", config.DefaultVersionFileLocation, "Location of the version file")
}

func runBumpCommand(cmd *cobra.Command, args []string) error {
	dryRunFlag, _ := cmd.Flags().GetBool("dry-run")
	versionFileFlag, _ := cmd.Flags().GetString("file")

	argVersionSegment := args[0]
	argVersionNumber := args[1]

	var versionInfo *config.VersionFile
	var err error

	if argVersionNumber != "" {
		versionInfo, err = config.ParseVersionNumber(argVersionNumber)
	} else {
		versionInfo, err = config.OpenVersionFile(versionFileFlag)
	}

	if err != nil {
		return err
	}

	switch argVersionSegment {
	case "major":
		versionInfo.Bump(config.BumpMajor)
	case "minor":
		versionInfo.Bump(config.BumpMinor)
	case "patch":
		versionInfo.Bump(config.BumpPatch)
	default:
		return fmt.Errorf("unknown version segment: %s", args[0])
	}

	if !dryRunFlag {
		if versionFileFlag != versionInfo.Path {
			versionInfo.Path = versionFileFlag
			log.Println(txt.Bluef("saving version file to \"%s\"", versionInfo.Path))
		}
		if err := versionInfo.Save(); err != nil {
			return err
		}
	}

	fmt.Printf("%s\n", versionInfo.String())

	return nil
}
