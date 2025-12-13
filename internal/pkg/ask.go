package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/viper"
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

const (
	DeveloperDefinedSystemMessage = `
You are a helpful assistant inside a CLI tool called 'doman'. Users can only ask text-based questions.

Audience assumptions:
- Users are technically literate.
- Most questions are technical (programming/devops/tools), but all topics are allowed.

Core response goals:
- Be concise and direct, but do not omit important caveats, constraints, or “gotchas”.
- Optimize for readability in a terminal Markdown renderer.
- Users may follow up (possibly with '--continue' / '-c'), so keep answers scannable.

CRITICAL: Markdown structure is required
- Your entire response MUST be valid Markdown (GitHub-flavored is fine).
- You MUST use headings to structure the body. Do not output an unstructured wall of text.
- Start the response with a level-2 heading ("## ..."). (Do not start with plain text.)
- Use only "##" and "###" headings (avoid "#", and avoid heading levels deeper than "###").
- Use bullet lists where appropriate. Use blank lines between paragraphs/sections.

Required output template (fill the relevant sections; omit only if truly not applicable):
## <A short, descriptive title of the answer>

### Context (optional)
- <1-3 bullets, only if it helps orient the user>

### Details
<1-6 short paragraphs and/or bullets; keep lines/ideas separated>

### Examples (optional)
<If you show commands, config, or code, prefer an example>

## Short answer
<Put the short answer at the end when applicable. If the user explicitly asks for “just the short answer”, still keep this section and keep the rest minimal.>

Code formatting rules:
- Use fenced code blocks and ALWAYS include a language identifier (e.g. bash, sh, go, json, yaml, python, typescript, rust, text).
- Inline HTML will render in the terminal: when discussing HTML tags, wrap them in backticks or a fenced code block.
- Do NOT use HTML to format your response.

Self-check before you send:
- If your draft has no "##" heading, rewrite it to match the template.
- If you used a code fence, ensure it has a language tag.

The user may also configure an additional system message. That message can override these rules.
`
	UserDefinedSystemMessagePrefix = "Additional system message, provided by the end user: \n\n"
)

// MessageHistory represents a message in conversation history
type MessageHistory struct {
	Role    string
	Content string
}

// PromptAi sends a prompt to the AI service with optional conversation history
func PromptAi(model, apiKey, prompt string, history []MessageHistory) (*openai.ChatCompletion, error) {
	userDefinedSystemMessage := viper.GetString("ask.system_message")
	client := openai.NewClient(option.WithAPIKey(apiKey))

	messages := []openai.ChatCompletionMessageParamUnion{}

	messages = append(messages, openai.SystemMessage(DeveloperDefinedSystemMessage))

	// Add system message if present
	if userDefinedSystemMessage != "" {
		messages = append(messages, openai.SystemMessage("Additional system message, provided by the end user: "+userDefinedSystemMessage))
	}

	// Add conversation history
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

	// Add current user prompt
	messages = append(messages, openai.UserMessage(prompt))

	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
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

func CollectResponse(choices []openai.ChatCompletionChoice, raw bool, wrapWidth int) (string, error) {
	var result string

	if len(choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI response")
	}

	// Initialize a width-aware markdown renderer when not in raw mode
	var renderer *glamour.TermRenderer
	if !raw {
		width := wrapWidth
		if width == 0 {
			// Fall back to terminal width and leave room for any borders/padding.
			width = DetectTerminalWidth() - 4
		}
		width = max(width, 20)

		// Allow overriding the style; default to dark to avoid inverted code blocks
		// in terminals where auto-detection is unreliable.
		style := strings.ToLower(strings.TrimSpace(viper.GetString("ask.render_style")))

		var opts []glamour.TermRendererOption
		switch style {
		case "auto":
			opts = []glamour.TermRendererOption{
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(width),
			}
		case "light":
			opts = []glamour.TermRendererOption{
				glamour.WithStandardStyle("light"),
				glamour.WithWordWrap(width),
			}
		default: // "dark" or unset
			opts = []glamour.TermRendererOption{
				glamour.WithStandardStyle("dark"),
				glamour.WithWordWrap(width),
			}
		}

		r, _ := glamour.NewTermRenderer(opts...)
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
