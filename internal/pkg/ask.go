package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	style, _ = glamour.NewTermRenderer(glamour.WithAutoStyle())
)

func PromptAi(model, apiKey, prompt string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are an assistant reached through a CLI tool. Please be helpful & concise in your responses. Use markdown formatting for structure and emphasis."),
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(1000),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get chat completion: %w", err)
	}

	// Process the chat completion response
	var result string

	for _, choice := range chatCompletion.Choices {
		content := choice.Message.Content
		content = strings.ReplaceAll(content, "\n\n\n\n", "\n\n")
		content = strings.TrimSpace(content)

		formatted, err := style.Render(content)
		if err != nil {
			return "", fmt.Errorf("failed to render response: %w", err)
		}

		result += formatted
	}

	return result, nil
}
