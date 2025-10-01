package cmd

import (
	"fmt"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

const shrugEmoticon = "¯\\_(ツ)_/¯"

var ShrugCommand = &cobra.Command{
	Use:   "shrug",
	Short: "Print a shrug emoticon " + shrugEmoticon,
	Run:   executeShrug,
}

func init() {
	ShrugCommand.Flags().BoolP("copy", "c", false, "Copy the shrug emoticon to clipboard")
}

func executeShrug(cmd *cobra.Command, args []string) {
	copy, err := cmd.Flags().GetBool("copy")
	if err != nil {
		fmt.Println("Error reading 'copy' flag:", err)
		return
	}

	if copy {
		err := pkg.CopyToClipboard(shrugEmoticon)
		if err != nil {
			fmt.Println("Error copying to clipboard:", err)
			return
		}
		pkg.PrintSuccess("Copied to clipboard, I guess %s", shrugEmoticon)

		return
	}

	fmt.Println(shrugEmoticon)
}
