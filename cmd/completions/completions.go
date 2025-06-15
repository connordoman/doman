package completions

import "github.com/spf13/cobra"

var CompletionsCommand = &cobra.Command{
	Use:   "completions",
	Short: "Generate shell completions for doman",
	Long:  "Commands for generating and setting up shell completions for the doman CLI.",
}

func init() {
	CompletionsCommand.AddCommand(generateCommand)
	CompletionsCommand.AddCommand(setupCommand)
}
