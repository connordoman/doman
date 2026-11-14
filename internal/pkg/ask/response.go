package ask

import (
	"fmt"
	"strings"

	"doman.sh/doman/internal/pkg"
	"github.com/charmbracelet/glamour"
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

// NormalizeContent tidies model output before it is stored or rendered.
func NormalizeContent(content string) string {
	return normalizeContent(content)
}

func normalizeContent(content string) string {
	content = strings.ReplaceAll(content, "\n\n\n\n", "\n\n")
	return strings.TrimSpace(content)
}
