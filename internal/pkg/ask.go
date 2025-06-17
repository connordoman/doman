package pkg

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func PromptAi(model, apiKey, prompt string) error {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("You are an assistant reached through a CLI tool. Please be helpful & concise in your responses. Please respond using ANSI coloring & format escape codes that appropriately supports your response. Please do not use markdown formatting."),
			openai.UserMessage(prompt),
		},
		MaxCompletionTokens: openai.Int(1000),
	})
	if err != nil {
		return fmt.Errorf("failed to get chat completion: %w", err)
	}

	// Process the chat completion response
	for _, choice := range chatCompletion.Choices {
		fmt.Printf("🤖 %s\n", choice.Message.Content)
	}

	return nil
}
