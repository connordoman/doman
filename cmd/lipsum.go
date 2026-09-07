package cmd

import (
	"fmt"

	"doman.sh/doman/internal/pkg/lipsum"
	"github.com/spf13/cobra"
)

var LoremIpsumCommand = &cobra.Command{
	Use:   "lipsum",
	Short: "Generate lorem ipsum text up to n words",
	RunE:  runLoremIpsumCommand,
}

func init() {
	rootCmd.AddCommand(LoremIpsumCommand)

	LoremIpsumCommand.Flags().IntP("words", "w", -1, "Number of words to print")
	LoremIpsumCommand.Flags().IntP("paragraphs", "p", 1, "Splits result into n paragraphs of random length")
}

func runLoremIpsumCommand(cmd *cobra.Command, args []string) error {
	wordsFlag, _ := cmd.Flags().GetInt("words")
	paragraphsFlag, _ := cmd.Flags().GetInt("paragraphs")

	opening := lipsum.GetOpening()

	result := opening
	if wordsFlag > lipsum.OpeningWordCount {
		result = lipsum.GenerateLoremIpsum(wordsFlag)
	}

	if paragraphsFlag > 1 {
		var err error
		result, err = lipsum.SplitNParagraphs(result, paragraphsFlag)
		if err != nil {
			return err
		}
	}

	fmt.Printf("%s\n", lipsum.StringifyTokens(result))

	return nil
}
