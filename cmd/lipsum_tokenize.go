package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/connordoman/doman/internal/pkg/lipsum"
	"github.com/spf13/cobra"
)

var LoremIpsumTokenizeCommand = &cobra.Command{
	Use:   "tokenize <text>",
	Short: "Tokenize string using lorem ipsum tokenizer",
	RunE:  runLoremIpsumTokenizeCommand,
	Args:  cobra.ExactArgs(1),
}

func init() {
	LoremIpsumCommand.AddCommand(LoremIpsumTokenizeCommand)
}

func runLoremIpsumTokenizeCommand(cmd *cobra.Command, args []string) error {
	text := args[0]

	tokens := lipsum.Tokenize(text)

	tokenJSON, err := json.MarshalIndent(tokens, "", " ")
	if err != nil {
		return fmt.Errorf("failed to convert tokens to JSON: %w", err)
	}

	fmt.Printf("%s\n", tokenJSON)

	return nil
}
