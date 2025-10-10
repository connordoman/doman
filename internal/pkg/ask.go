package pkg

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// Cost per million tokens
type ModelPricing struct {
	InputCost       float64
	CachedInputCost float64
	OutputCost      float64
}

var (
	costTable = map[string]ModelPricing{
		"gpt-5-nano": {
			InputCost:       0.050,
			CachedInputCost: 0.005,
			OutputCost:      0.400,
		},
		"gpt-5-mini": {
			InputCost:       0.250,
			CachedInputCost: 0.025,
			OutputCost:      2.000,
		},
		"gpt-5": {
			InputCost:       1.250,
			CachedInputCost: 0.125,
			OutputCost:      10.000,
		},
	}
)

var AskSplashText = []string{
	"Talking to robots",
	"Getting the skinny",
	"Finding the hay in the needle stack",
	"Pushing to production",
	"Talking to the little man",
	"Getting a second opinion",
	"Looking it up for you (even though you could probably do it)",
	"Getting the lowdown",
	"Writing a strongly worded letter",
	"Lighting a signal fire",
	"Using a small village's electricity budget for this",
	"Asking the AI to ask the AI",
	"Preparing to shut off PS4",
	"Castling",
	"Throwing it back",
	"Consulting the tea leaves",
	"Praying to god",
	"Shooting the messenger",
	"Texting my mom",
	"Figuring out the hard way",
	"Firing my assistant",
	"Thinking about stuff",
	"Letting the voices in",
	"Getting the instructions out of the garbage",
	"Going back to school",
	"Definitely not just Googling it",
	"Finishing my protein shake first",
	"Playing the long game",
	"Asking my supervisor",
	"F***ing around in hope of finding out",
	"Calling your boss",
	"Reading your diary",
	"Waking up early to get the worm",
	"Making a quick buck",
	"Sending a message in a bottle",
	"Asking Dr. Wilson for a consult",
	"I'm working on it",
	"Swiping right",
	"Finishing my bathroom break",
}

func PromptAi(model, apiKey, prompt string) (*openai.ChatCompletion, error) {
	systemMessage := viper.GetString("ask.system_message")
	client := openai.NewClient(option.WithAPIKey(apiKey))
	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMessage),
			openai.UserMessage(prompt),
		},
		// MaxCompletionTokens: openai.Int(2000), // Increased from 1000
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get chat completion: %w", err)
	}

	// Debug: print usage information
	// fmt.Printf("DEBUG: Usage - Prompt tokens: %d, Completion tokens: %d, Total tokens: %d\n",
	// 	chatCompletion.Usage.PromptTokens, chatCompletion.Usage.CompletionTokens, chatCompletion.Usage.TotalTokens)

	return chatCompletion, nil
}

func CollectResponse(choices []openai.ChatCompletionChoice, raw bool) (string, error) {
	var result string

	if len(choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI response")
	}

	// Initialize a width-aware markdown renderer when not in raw mode
	var renderer *glamour.TermRenderer
	if !raw {
		width := detectTerminalWidth()
		if width < 20 {
			width = 20
		}
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(width),
		)
		renderer = r
	}

	for i, choice := range choices {
		// Debug: print the choice structure
		// fmt.Printf("DEBUG: Choice %d - Role: %s, FinishReason: %s\n",
		// 	i, choice.Message.Role, choice.FinishReason)

		content := choice.Message.Content
		if content == "" {
			return "", fmt.Errorf("received empty content from AI response (choice %d, finish_reason: %s)", i, choice.FinishReason)
		}

		content = strings.ReplaceAll(content, "\n\n\n\n", "\n\n")
		content = strings.TrimSpace(content)

		if raw {
			result += content
		} else {
			formatted, err := renderer.Render(content)
			if err != nil {
				return "", fmt.Errorf("failed to render response: %w", err)
			}

			result += formatted
		}
	}

	if raw {
		result += "\n"
	}

	return result, nil
}

func AskingSpinner(prompt string, actionWithError func(ctx context.Context) error) *spinner.Spinner {
	return spinner.New().Title(prompt).Style(lipgloss.NewStyle().Foreground(lipgloss.Color("#2563eb"))).ActionWithErr(actionWithError)
}

func CalculateCost(model string, completion *openai.ChatCompletion) (float64, bool) {
	pricing, exists := costTable[model]
	if !exists {
		return 0, false
	}

	var totalCost float64

	inputTokens := float64(completion.Usage.PromptTokens)
	cachedTokens := float64(completion.Usage.PromptTokensDetails.CachedTokens)
	outputTokens := float64(completion.Usage.CompletionTokens)

	totalCost += (inputTokens - cachedTokens) * pricing.InputCost
	totalCost += cachedTokens * pricing.CachedInputCost
	totalCost += outputTokens * pricing.OutputCost

	return totalCost / 1_000_000, true
}

// detectTerminalWidth returns the current terminal width in columns.
// If width cannot be determined, it falls back to a sensible default of 80.
func detectTerminalWidth() int {
	// Respect COLUMNS if set and valid
	if colsStr := os.Getenv("COLUMNS"); colsStr != "" {
		if cols, err := strconv.Atoi(colsStr); err == nil && cols > 0 {
			return cols
		}
	}

	// Try stdout
	if term.IsTerminal(int(os.Stdout.Fd())) {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			return w
		}
	}

	// Try stderr
	if term.IsTerminal(int(os.Stderr.Fd())) {
		if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 {
			return w
		}
	}

	return 80
}
