package ask

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"doman.sh/doman/internal/pkg"
	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/muesli/termenv"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// documentMargin is the left margin glamour's standard styles indent by.
const documentMargin = 2

// MarkdownRenderer renders Markdown for the terminal with glamour, adding
// support for GitHub alerts (> [!NOTE] and friends), which glamour would
// otherwise print as plain blockquotes.
type MarkdownRenderer struct {
	doc   *glamour.TermRenderer
	style glamouransi.StyleConfig
	light bool
	width int

	// alert renders alert bodies; it is built on first use since most
	// answers contain none.
	alert *glamour.TermRenderer
}

// NewMarkdownRenderer builds a renderer honouring the configured
// ask.render_style. A wrapWidth of 0 falls back to the detected terminal width.
func NewMarkdownRenderer(wrapWidth int) (*MarkdownRenderer, error) {
	width := wrapWidth
	if width == 0 {
		width = pkg.DetectTerminalWidth() - 4
	}
	width = max(width, 20)

	style, light := resolveRenderStyle(viper.GetString("ask.render_style"))

	doc, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	return &MarkdownRenderer{doc: doc, style: style, light: light, width: width}, nil
}

// Render formats content, drawing any GitHub alerts as titled boxes.
func (r *MarkdownRenderer) Render(content string) (string, error) {
	if !hasAlerts(content) {
		return r.doc.Render(content)
	}

	segments := splitAlerts(content)
	if len(segments) == 1 && segments[0].kind == "" {
		return r.doc.Render(content)
	}

	margin := strings.Repeat(" ", documentMargin)
	boxWidth := r.width - documentMargin

	blocks := make([]string, 0, len(segments))
	for _, seg := range segments {
		if seg.kind == "" {
			rendered, err := r.doc.Render(seg.body)
			if err != nil {
				return "", err
			}
			if rendered = trimBlankLines(rendered); rendered != "" {
				blocks = append(blocks, rendered)
			}
			continue
		}

		body := ""
		if strings.TrimSpace(seg.body) != "" {
			alert, err := r.alertRenderer(boxWidth - 4)
			if err != nil {
				return "", err
			}
			rendered, err := alert.Render(seg.body)
			if err != nil {
				return "", err
			}
			body = trimBlankLines(rendered)
		}

		box := renderAlertBox(alertKinds[seg.kind], body, boxWidth, r.light)
		blocks = append(blocks, margin+strings.ReplaceAll(box, "\n", "\n"+margin))
	}

	// Match glamour's own framing: a leading newline and a trailing blank line.
	return "\n" + strings.Join(blocks, "\n\n") + "\n\n", nil
}

func (r *MarkdownRenderer) alertRenderer(width int) (*glamour.TermRenderer, error) {
	if r.alert != nil {
		return r.alert, nil
	}

	// The box supplies the indentation, so drop the document margin.
	style := r.style
	style.Document.Margin = new(uint)

	alert, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(max(width, 10)),
	)
	if err != nil {
		return nil, err
	}
	r.alert = alert
	return alert, nil
}

var (
	autoStyleOnce  sync.Once
	autoStyleLight bool
	autoStyleNoTTY bool
)

// resolveRenderStyle maps the configured render style to a glamour style
// config, and reports whether it is a light theme so alert colours can match.
func resolveRenderStyle(name string) (glamouransi.StyleConfig, bool) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "light":
		return styles.LightStyleConfig, true
	case "auto":
		// Querying the terminal background is slow and noisy inside a TUI,
		// so do it once per process.
		autoStyleOnce.Do(func() {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				autoStyleNoTTY = true
				return
			}
			autoStyleLight = !termenv.HasDarkBackground()
		})
		if autoStyleNoTTY {
			return styles.NoTTYStyleConfig, false
		}
		if autoStyleLight {
			return styles.LightStyleConfig, true
		}
		return styles.DarkStyleConfig, false
	default:
		return styles.DarkStyleConfig, false
	}
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
