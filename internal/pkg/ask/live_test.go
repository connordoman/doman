package ask

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestParseQuickCommand(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantPrompt string
		wantQuick  bool
	}{
		{"no command", "how do I rebase?", "how do I rebase?", false},
		{"leading", "/quick how do I rebase?", "how do I rebase?", true},
		{"trailing", "how do I rebase? /quick", "how do I rebase?", true},
		{"middle", "how do I /quick rebase?", "how do I rebase?", true},
		{"uppercase", "/QUICK how do I rebase?", "how do I rebase?", true},
		{"repeated", "/quick /quick how do I rebase?", "how do I rebase?", true},
		{"only the command", "/quick", "", true},
		{"not a word boundary", "/quicksort in go", "/quicksort in go", false},
		{"inside a path", "run cmd/quick now", "run cmd/quick now", false},
		{"keeps line breaks", "explain this:\n/quick\nfoo bar", "explain this:\nfoo bar", true},
		{"empty", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prompt, quick := ParseQuickCommand(tc.input)
			if prompt != tc.wantPrompt {
				t.Errorf("prompt = %q, want %q", prompt, tc.wantPrompt)
			}
			if quick != tc.wantQuick {
				t.Errorf("quick = %v, want %v", quick, tc.wantQuick)
			}
		})
	}
}

func TestQuickTurnSystemMessageIsScopedToTheTurn(t *testing.T) {
	if quickTurnSystemMessage(false) != "" {
		t.Fatal("normal turns should not carry the quick instruction")
	}
	if quickTurnSystemMessage(true) != LiveQuickTurnSystemMessage {
		t.Fatal("quick turns should carry the quick instruction")
	}
}

func TestAPIHistoryDropsNoticesAndPendingTurn(t *testing.T) {
	m := DefaultLiveChatModel()
	m.messages = []LiveMessage{
		{Role: LiveRoleUser, Content: "first"},
		{Role: LiveRoleAssistant, Content: "answer"},
		{Role: LiveRoleNotice, Content: "Request cancelled."},
		{Role: LiveRoleError, Content: "boom"},
		{Role: LiveRoleUser, Content: "second"},
	}

	history := m.apiHistory()

	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d: %+v", len(history), history)
	}
	if history[0].Content != "first" || history[1].Content != "answer" {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestSanitizeStoredContentStripsEscapeCodes(t *testing.T) {
	got := SanitizeStoredContent("\x1b[1mBold\x1b[0m answer\n")
	if got != "Bold answer" {
		t.Fatalf("got %q", got)
	}
}

// The transcript is re-rendered on resize, so it must never exceed the width it
// was given.
func TestTranscriptReflowsToWidth(t *testing.T) {
	m := DefaultLiveChatModel()
	m.cfg.Raw = true
	m.messages = []LiveMessage{
		{Role: LiveRoleUser, Content: strings.Repeat("word ", 60)},
		{Role: LiveRoleAssistant, Content: strings.Repeat("reply ", 60), Model: "gpt-4o-mini"},
	}

	for _, width := range []int{30, 60, 120} {
		for _, line := range strings.Split(m.renderTranscript(width), "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("width %d: line of %d runes: %q", width, got, line)
			}
		}
	}
}

func TestViewRendersAtSeveralSizes(t *testing.T) {
	chat := DefaultLiveChatModel()
	chat.messages = []LiveMessage{
		{Role: LiveRoleUser, Content: strings.Repeat("question ", 20)},
		{Role: LiveRoleAssistant, Content: "### Answer\n\n" + strings.Repeat("reply ", 40), Model: "gpt-4o-mini", Cost: 0.0001},
		{Role: LiveRoleNotice, Content: "Request cancelled."},
	}

	var model tea.Model = chat

	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 24}, {Width: 40, Height: 12}, {Width: 200, Height: 60}} {
		model, _ = model.Update(size)

		view := model.View()
		if !view.AltScreen {
			t.Fatal("live chat should run in the alternate screen")
		}
		if strings.TrimSpace(view.Content) == "" {
			t.Fatalf("empty view at %dx%d", size.Width, size.Height)
		}
		if got := strings.Count(view.Content, "\n") + 1; got > size.Height {
			t.Fatalf("view is %d lines tall at height %d", got, size.Height)
		}
	}
}

func TestSubmitOnlyQuickCommandDoesNotSend(t *testing.T) {
	m := DefaultLiveChatModel()
	m.width, m.height, m.ready = 80, 24, true
	m.input.SetValue("/quick")

	if cmd := m.submit(); cmd != nil {
		t.Fatal("a bare /quick should not start a request")
	}
	if m.waiting {
		t.Fatal("a bare /quick should not put the chat in a waiting state")
	}
	if len(m.messages) != 1 || m.messages[0].Role != LiveRoleNotice {
		t.Fatalf("expected a single notice, got %+v", m.messages)
	}
}

// The composer grows with its content, but never so far that the view outgrows
// the terminal.
func TestViewFitsTerminalWithATallComposer(t *testing.T) {
	sizes := []tea.WindowSizeMsg{
		{Width: 80, Height: 40},
		{Width: 80, Height: 12},
		{Width: 80, Height: 8},
		{Width: 30, Height: 10},
	}

	for _, size := range sizes {
		var model tea.Model = DefaultLiveChatModel()
		model, _ = model.Update(size)

		for i := range 12 {
			model, _ = model.Update(tea.PasteMsg{Content: fmt.Sprintf("line %d\n", i)})
		}

		chat, ok := model.(LiveChatModel)
		if !ok {
			t.Fatalf("unexpected model type %T", model)
		}
		if chat.input.Height() < 2 && size.Height > 10 {
			t.Errorf("%dx%d: composer did not grow with its content", size.Width, size.Height)
		}

		content := model.View().Content
		if got := strings.Count(content, "\n") + 1; got > size.Height {
			t.Errorf("%dx%d: view is %d lines tall", size.Width, size.Height, got)
		}
		for _, line := range strings.Split(content, "\n") {
			if got := lipgloss.Width(line); got > size.Width {
				t.Errorf("%dx%d: line is %d cells wide: %q", size.Width, size.Height, got, line)
			}
		}
	}
}
