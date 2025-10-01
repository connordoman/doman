package cmd

import (
	"fmt"
	"strconv"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var RandCommand = &cobra.Command{
	Use:   "rand <length>",
	Short: "Generate a random string of the specified length",
	Long:  `Generate a random string of the specified length. The string will consist of alphanumeric characters.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRandCommand,
}

func init() {
	RandCommand.Flags().BoolP("base64", "B", false, "Generate a base64 encoded random string")
	RandCommand.Flags().BoolP("hex", "H", false, "Generate a hex encoded random string")

	RandCommand.Flags().BoolP("url-safe", "U", false, "Ensure generated string is URL-safe")

	RandCommand.MarkFlagsMutuallyExclusive("base64", "hex")
	RandCommand.MarkFlagsMutuallyExclusive("hex", "url-safe")
}

func runRandCommand(cmd *cobra.Command, args []string) error {
	length, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	flagBase64, _ := cmd.Flags().GetBool("base64")
	flagHex, _ := cmd.Flags().GetBool("hex")
	flagURLSafe, _ := cmd.Flags().GetBool("url-safe")

	randFunction := pkg.RandomString
	if flagBase64 {
		if flagURLSafe {
			randFunction = pkg.RandomBase64URL
		} else {
			randFunction = pkg.RandomBase64
		}
	} else if flagHex {
		randFunction = pkg.RandomHex
	}

	randomString, err := randFunction(length)
	if err != nil {
		return err
	}

	fmt.Println(randomString)
	return nil
}
