package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/connordoman/doman/internal/pkg/ask"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var AskTitleCommand = &cobra.Command{
	Use:   "title",
	Short: "Test generating titles for a conversation",
	RunE:  runAskTitlesCommand,
}

func init() {
	AskCommand.AddCommand(AskTitleCommand)
}

func runAskTitlesCommand(cmd *cobra.Command, args []string) error {
	prompt := ""

	if len(args) > 0 {
		prompt = strings.TrimSpace(strings.Join(args, " "))
	} else {

		err := huh.NewInput().
			Title("Enter your question").
			Value(&prompt).
			Run()
		if err != nil {
			return fmt.Errorf("failed to get user input: %w", err)
		}
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	fmt.Println(ask.FormatPrompt(prompt, false))

	apiKey := viper.GetString("ask.openai.api_key")
	if apiKey == "" {
		return fmt.Errorf("API Key is required, please set it using --api-key or environment variable OPENAI_API_KEY")
	}
	titleModel := viper.GetString("ask.title_model")
	if titleModel == "" {
		titleModel = "gpt-5-nano"
	}

	title, err := ask.GenerateShortTitle(cmd.Context(), apiKey, titleModel, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate short title: %w", err)
	}

	fmt.Printf("Title: %s\n", txt.Boldf("%s", title))
	return nil
}
