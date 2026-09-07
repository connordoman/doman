package ask

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/x/ansi"
	"github.com/connordoman/windy"
)

const liveSeparator = " • "

// renderTranscript renders every message at the given width. It builds one
// Markdown renderer for the whole pass so a resize does not pay for a renderer
// per message.
func (m LiveChatModel) renderTranscript(width int) string {
	if len(m.messages) == 0 {
		return m.emptyTranscriptView(width)
	}

	var renderer *glamour.TermRenderer
	if !m.cfg.Raw {
		// Glamour indents its own output, so leave room for that margin.
		renderer, _ = NewMarkdownRenderer(max(width-2, 20))
	}

	blocks := make([]string, 0, len(m.messages))
	for _, msg := range m.messages {
		blocks = append(blocks, m.renderMessage(msg, width, renderer))
	}

	return strings.Join(blocks, "\n")
}

func (m LiveChatModel) renderMessage(msg LiveMessage, width int, renderer *glamour.TermRenderer) string {
	switch msg.Role {
	case LiveRoleUser:
		return m.renderUserMessage(msg, width)
	case LiveRoleAssistant:
		return m.renderAssistantMessage(msg, width, renderer)
	case LiveRoleError:
		return m.renderNotice("Error", msg.Content, windy.Red400, width)
	default:
		return m.renderNotice("", msg.Content, windy.Neutral500, width)
	}
}

func (m LiveChatModel) renderUserMessage(msg LiveMessage, width int) string {
	accent := windy.Blue400
	label := "You"
	if msg.Quick {
		accent = windy.Yellow400
		label = "⚡ You"
	}

	// Border column plus a padding column.
	bodyWidth := max(width-2, 1)

	block := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(liveColor(accent)).Render(label),
		lipgloss.NewStyle().Width(bodyWidth).Render(msg.Content),
	)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(liveColor(accent)).
		PaddingLeft(1).
		Render(block) + "\n"
}

func (m LiveChatModel) renderAssistantMessage(msg LiveMessage, width int, renderer *glamour.TermRenderer) string {
	body := msg.Content

	if renderer != nil {
		if rendered, err := renderer.Render(msg.Content); err == nil {
			body = strings.Trim(rendered, "\n")
		}
	} else {
		body = lipgloss.NewStyle().Width(max(width-2, 1)).PaddingLeft(2).Render(msg.Content)
	}

	meta := []string{"ChatGPT"}
	if msg.Model != "" {
		meta = append(meta, msg.Model)
	}
	if msg.Cost > 0 {
		meta = append(meta, fmt.Sprintf("$%.5f", msg.Cost))
	}
	if msg.Elapsed > 0 {
		meta = append(meta, formatLiveDuration(msg.Elapsed))
	}
	if msg.Quick {
		meta = append(meta, "quick")
	}

	footer := lipgloss.NewStyle().
		Foreground(liveColor(windy.Neutral500)).
		PaddingLeft(2).
		MaxWidth(width).
		Render(strings.Join(meta, liveSeparator))

	return lipgloss.JoinVertical(lipgloss.Left, body, footer) + "\n"
}

func (m LiveChatModel) renderNotice(label, content string, accent windy.TailwindColor, width int) string {
	text := content
	if label != "" {
		text = label + ": " + content
	}

	return lipgloss.NewStyle().
		Foreground(liveColor(accent)).
		Italic(true).
		PaddingLeft(2).
		Width(width).
		Render(text) + "\n"
}

func (m LiveChatModel) emptyTranscriptView(width int) string {
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(liveColor(windy.Blue400)).Render("A new conversation"),
		"",
		"Ask a question below and press enter.",
		fmt.Sprintf("Include %s anywhere in a message for a short answer from %s.", QuickCommand, m.quickModel()),
	}

	return lipgloss.NewStyle().
		Foreground(liveColor(windy.Neutral500)).
		PaddingLeft(2).
		PaddingTop(1).
		Width(width).
		Render(strings.Join(lines, "\n"))
}

func (m LiveChatModel) headerView() string {
	title := m.title
	if !IsMeaningfulTitle(title) {
		title = shortConversationID(m.cfg.ConversationID)
	}

	left := lipgloss.NewStyle().Bold(true).Foreground(liveColor(windy.Blue400)).Render("doman ask") +
		lipgloss.NewStyle().Foreground(liveColor(windy.Neutral500)).Render(liveSeparator+title)

	meta := []string{m.cfg.Model}
	if m.totalCost > 0 {
		meta = append(meta, fmt.Sprintf("$%.5f", m.totalCost))
	}
	right := lipgloss.NewStyle().Foreground(liveColor(windy.Neutral500)).Render(strings.Join(meta, liveSeparator))

	width := m.contentWidth()
	gutter := lipgloss.NewStyle().PaddingLeft(1)

	rule := lipgloss.NewStyle().
		Foreground(liveColor(windy.Neutral700)).
		Render(strings.Repeat("─", max(width, 1)))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		gutter.Render(m.spread(left, right, width)),
		gutter.Render(rule),
	)
}

func (m LiveChatModel) statusView() string {
	padding := lipgloss.NewStyle().PaddingLeft(1)

	if m.waiting {
		elapsed := formatLiveDuration(time.Since(m.startedAt))
		status := fmt.Sprintf("%s %s... %s", m.spinner.View(), m.splash, elapsed)
		return padding.
			Foreground(liveColor(windy.Neutral400)).
			MaxWidth(m.width).
			Render(status)
	}

	hint := "Ready" + liveSeparator + QuickCommand + " anywhere in a message for a fast answer"
	return padding.
		Foreground(liveColor(windy.Neutral600)).
		MaxWidth(m.width).
		Render(hint)
}

func (m LiveChatModel) inputView() string {
	accent := windy.Neutral700
	if m.waiting {
		accent = windy.Neutral800
	} else if _, quick := ParseQuickCommand(m.input.Value()); quick {
		accent = windy.Yellow400
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(liveColor(accent)).
		Padding(0, 1).
		Width(m.contentWidth()).
		Render(m.input.View())
}

func (m LiveChatModel) helpView() string {
	keys := []string{
		"enter send",
		"alt+enter newline",
		"pgup/pgdn scroll",
	}
	if m.waiting {
		keys = append(keys, "esc cancel")
	}
	keys = append(keys, "ctrl+c quit")

	return lipgloss.NewStyle().
		Foreground(liveColor(windy.Neutral600)).
		PaddingLeft(1).
		MaxWidth(m.width).
		Render(strings.Join(keys, liveSeparator))
}

func (m LiveChatModel) windowTitle() string {
	if IsMeaningfulTitle(m.title) {
		return "doman ask: " + m.title
	}
	return "doman live"
}

// spread pushes right up against the far edge of width, truncating left first
// when the terminal is too narrow to hold both.
func (m LiveChatModel) spread(left, right string, width int) string {
	if width <= 0 {
		return left
	}

	rightWidth := lipgloss.Width(right)
	if rightWidth+2 > width {
		return ansi.Truncate(left, width, "…")
	}

	available := width - rightWidth - 1
	left = ansi.Truncate(left, available, "…")
	gap := max(width-lipgloss.Width(left)-rightWidth, 1)

	return left + strings.Repeat(" ", gap) + right
}

func formatLiveDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return d.Round(time.Second).String()
}

func shortConversationID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
