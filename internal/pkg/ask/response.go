package ask

import (
	"fmt"
	"strings"

	"doman.sh/internal/pkg"
	"github.com/charmbracelet/glamour"
	openai "github.com/openai/openai-go"
	"github.com/spf13/viper"
)

// NewMarkdownRenderer builds a glamour renderer honouring the configured
// ask.render_style. A wrapWidth of 0 falls back to the detected terminal width.
func NewMarkdownRenderer(wrapWidth int) (*glamour.TermRenderer, error) {
	width := wrapWidth
	if width == 0 {
		width = pkg.DetectTerminalWidth() - 4
	}
	width = max(width, 20)

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
	default:
		opts = []glamour.TermRendererOption{
			glamour.WithStandardStyle("dark"),
			glamour.WithWordWrap(width),
		}
	}

	return glamour.NewTermRenderer(opts...)
}

// RenderMarkdown formats content for the terminal. When raw is set the content
// is returned untouched apart from whitespace normalisation.
func RenderMarkdown(content string, raw bool, wrapWidth int) (string, error) {
	content = normalizeContent(content)

	if raw {
		return content, nil
	}

	renderer, err := NewMarkdownRenderer(wrapWidth)
	if err != nil {
		return "", fmt.Errorf("failed to create Markdown renderer: %w", err)
	}

	formatted, err := renderer.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render response: %w", err)
	}

	return formatted, nil
}

func CollectResponse(choices []openai.ChatCompletionChoice, raw bool, wrapWidth int) (string, error) {
	var result string

	if len(choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI response")
	}

	var renderer *glamour.TermRenderer
	if !raw {
		r, err := NewMarkdownRenderer(wrapWidth)
		if err != nil {
			return "", fmt.Errorf("failed to create Markdown renderer: %w", err)
		}
		renderer = r
	}

	for i, choice := range choices {
		content := choice.Message.Content
		if content == "" {
			return "", fmt.Errorf("received empty content from AI response (choice %d, finish_reason: %s)", i, choice.FinishReason)
		}

		content = normalizeContent(content)

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

// CollectRawResponse joins the raw Markdown of every choice without rendering
// it, so callers can store the model output and render it later.
func CollectRawResponse(choices []openai.ChatCompletionChoice) (string, error) {
	if len(choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI response")
	}

	parts := make([]string, 0, len(choices))
	for i, choice := range choices {
		content := choice.Message.Content
		if content == "" {
			return "", fmt.Errorf("received empty content from AI response (choice %d, finish_reason: %s)", i, choice.FinishReason)
		}

		parts = append(parts, normalizeContent(content))
	}

	return strings.Join(parts, "\n\n"), nil
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\n\n\n\n", "\n\n")
	return strings.TrimSpace(content)
}
