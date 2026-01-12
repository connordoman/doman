package ask

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/connordoman/doman/internal/pkg"
	openai "github.com/openai/openai-go"
	"github.com/spf13/viper"
)

func CollectResponse(choices []openai.ChatCompletionChoice, raw bool, wrapWidth int) (string, error) {
	var result string

	if len(choices) == 0 {
		return "", fmt.Errorf("no choices returned from AI response")
	}

	var renderer *glamour.TermRenderer
	if !raw {
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

		r, _ := glamour.NewTermRenderer(opts...)
		renderer = r
	}

	for i, choice := range choices {
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
