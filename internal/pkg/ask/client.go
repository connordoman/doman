package ask

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/viper"
)

// PromptOptions describes a single chat completion request.
type PromptOptions struct {
	Model   string
	APIKey  string
	Prompt  string
	History []MessageHistory

	// SystemMessage is the conversation-level system message. It is sent at
	// the head of every request.
	SystemMessage string

	// TurnSystemMessage is injected immediately before the user prompt and is
	// never written back into History, so per-turn instructions (quick mode,
	// for example) cost tokens only on the turn that actually uses them.
	TurnSystemMessage string
}

func PromptAI(model, apiKey, prompt string, history []MessageHistory) (*openai.ChatCompletion, error) {
	return PromptAIWithSystemMessage(model, apiKey, prompt, history, DeveloperDefinedSystemMessage)
}

func PromptAIWithSystemMessage(model, apiKey, prompt string, history []MessageHistory, systemMessage string) (*openai.ChatCompletion, error) {
	return PromptAIWithOptions(context.Background(), PromptOptions{
		Model:         model,
		APIKey:        apiKey,
		Prompt:        prompt,
		History:       history,
		SystemMessage: systemMessage,
	})
}

func PromptAIWithOptions(ctx context.Context, opts PromptOptions) (*openai.ChatCompletion, error) {
	userDefinedSystemMessage := viper.GetString("ask.system_message")
	preferredLanguages := readPreferredLanguages()
	client := openai.NewClient(option.WithAPIKey(opts.APIKey))

	messages := []openai.ChatCompletionMessageParamUnion{}

	if opts.SystemMessage != "" {
		systemMessage := opts.SystemMessage
		if len(preferredLanguages) > 0 {
			systemMessage = fmt.Sprintf("%s\n\nPreferred programming languages: %s.", systemMessage, strings.Join(preferredLanguages, ", "))
		}
		messages = append(messages, openai.SystemMessage(systemMessage))
	}

	if userDefinedSystemMessage != "" {
		messages = append(messages, openai.SystemMessage(UserDefinedSystemMessagePrefix+userDefinedSystemMessage))
	}

	for _, msg := range opts.History {
		switch msg.Role {
		case "system":
			messages = append(messages, openai.SystemMessage(msg.Content))
		case "user":
			messages = append(messages, openai.UserMessage(msg.Content))
		case "assistant":
			messages = append(messages, openai.AssistantMessage(msg.Content))
		}
	}

	if opts.TurnSystemMessage != "" {
		messages = append(messages, openai.SystemMessage(opts.TurnSystemMessage))
	}

	messages = append(messages, openai.UserMessage(opts.Prompt))

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    opts.Model,
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
