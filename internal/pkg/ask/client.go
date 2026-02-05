package ask

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/viper"
)

func PromptAI(model, apiKey, prompt string, history []MessageHistory) (*openai.ChatCompletion, error) {
	return PromptAIWithSystemMessage(model, apiKey, prompt, history, DeveloperDefinedSystemMessage)
}

func PromptAIWithSystemMessage(model, apiKey, prompt string, history []MessageHistory, systemMessage string) (*openai.ChatCompletion, error) {
	userDefinedSystemMessage := viper.GetString("ask.system_message")
	preferredLanguages := readPreferredLanguages()
	client := openai.NewClient(option.WithAPIKey(apiKey))

	messages := []openai.ChatCompletionMessageParamUnion{}

	if systemMessage != "" {
		if len(preferredLanguages) > 0 {
			systemMessage = fmt.Sprintf("%s\n\nPreferred programming languages: %s.", systemMessage, strings.Join(preferredLanguages, ", "))
		}
		messages = append(messages, openai.SystemMessage(systemMessage))
	}

	if userDefinedSystemMessage != "" {
		messages = append(messages, openai.SystemMessage("Additional system message, provided by the end user: "+userDefinedSystemMessage))
	}

	for _, msg := range history {
		switch msg.Role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	messages = append(messages, openai.UserMessage(prompt))

	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get chat completion: %w", err)
	}

	return chatCompletion, nil
}

func readPreferredLanguages() []string {
	languages := viper.GetStringSlice("ask.preferred_languages")
	if len(languages) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(languages))
	for _, language := range languages {
		language = strings.TrimSpace(language)
		if language != "" {
			normalized = append(normalized, language)
		}
	}

	return normalized
}

func GenerateShortTitle(ctx context.Context, apiKey, model, prompt string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You write short, 3-8 word titles for user prompts. Respond with a concise title only, no quotes or markdown. Keep it under 80 characters."),
		openai.UserMessage(prompt),
	}

	completion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:       model,
		Messages:    messages,
		Temperature: openai.Float(1),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate short title: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no title choices returned")
	}

	title := strings.TrimSpace(completion.Choices[0].Message.Content)
	title = strings.Trim(title, `"'""`)
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.Join(strings.Fields(title), " ")

	return truncateTitle(title, maxConversationTitleLength), nil
}
