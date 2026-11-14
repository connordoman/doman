package ask

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/spf13/viper"
)

// PromptOptions describes a single turn sent through a Session.
type PromptOptions struct {
	Model   string
	APIKey  string
	Prompt  string
	History []MessageHistory

	// SystemMessage is the conversation-level system message. It is sent at
	// the head of every request.
	SystemMessage string

	// TurnSystemMessage is added to this turn's instructions only and is never
	// written back into History, so per-turn instructions (quick mode, for
	// example) cost tokens only on the turn that actually uses them.
	TurnSystemMessage string

	// Settings are the Responses API options (reasoning, verbosity, ...).
	Settings ResponseSettings
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
