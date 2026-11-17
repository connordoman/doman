package ask

import (
	"os"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestMain(m *testing.M) {
	copyToClipboard = func(string) error { return nil }
	os.Exit(m.Run())
}

// selectionChat returns a sized chat whose transcript holds known plain lines.
func selectionChat(t *testing.T) LiveChatModel {
	t.Helper()

	m := DefaultLiveChatModel()
	m.cfg.Raw = true
	m.messages = []LiveMessage{{Role: LiveRoleNotice, Content: "hello world"}}

	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	chat := model.(LiveChatModel)
	chat.viewport.GotoTop()
	return chat
}

// screenOf returns the screen cell that shows the given transcript text.
func screenOf(t *testing.T, m LiveChatModel, needle string) (x, y int) {
	t.Helper()
	for i, line := range m.transcriptLines {
		if col := strings.Index(ansi.Strip(line), needle); col >= 0 {
			return col + liveTranscriptLeft, i - m.viewport.YOffset() + liveHeaderHeight
		}
	}
	t.Fatalf("%q not found in transcript", needle)
	return 0, 0
}

func drag(m LiveChatModel, x0, y0, x1, y1 int) LiveChatModel {
	var model tea.Model = m
	model, _ = model.Update(tea.MouseClickMsg{X: x0, Y: y0, Button: tea.MouseLeft})
	model, _ = model.Update(tea.MouseMotionMsg{X: x1, Y: y1, Button: tea.MouseLeft})
	model, _ = model.Update(tea.MouseReleaseMsg{X: x1, Y: y1, Button: tea.MouseLeft})
	return model.(LiveChatModel)
}

func TestDragSelectionIncludesBothEnds(t *testing.T) {
	m := selectionChat(t)
	x, y := screenOf(t, m, "world")

	// Drag from 'w' to 'r': both the clicked cell and the cell under the
	// pointer belong to the selection.
	m = drag(m, x, y, x+2, y)
	if got := m.selectedText(); got != "wor" {
		t.Fatalf("selected %q, want %q", got, "wor")
	}
	if !strings.HasPrefix(m.flashText, "Copied") {
		t.Fatalf("expected a copy confirmation, got %q", m.flashText)
	}
}

func TestDragSelectionBackwards(t *testing.T) {
	m := selectionChat(t)
	x, y := screenOf(t, m, "world")

	m = drag(m, x+2, y, x, y)
	if got := m.selectedText(); got != "wor" {
		t.Fatalf("selected %q, want %q", got, "wor")
	}
}

func TestClickWithoutDragClearsSelection(t *testing.T) {
	m := selectionChat(t)
	x, y := screenOf(t, m, "world")

	var model tea.Model = m
	model, _ = model.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	model, _ = model.Update(tea.MouseReleaseMsg{X: x, Y: y, Button: tea.MouseLeft})

	if model.(LiveChatModel).selection.active {
		t.Fatal("a plain click should not leave a selection behind")
	}
}

func TestSelectionHighlightKeepsLineWidth(t *testing.T) {
	m := selectionChat(t)
	x, y := screenOf(t, m, "hello")
	m = drag(m, x, y, x+7, y)

	if !m.selection.active {
		t.Fatal("selection should survive the release so it stays visible")
	}

	before := m.viewport.View()
	after := m.highlightSelection(before, m.viewport.YOffset())
	if before == after {
		t.Fatal("selection was not highlighted")
	}

	beforeRows, afterRows := strings.Split(before, "\n"), strings.Split(after, "\n")
	for i := range beforeRows {
		if ansi.Strip(beforeRows[i]) != ansi.Strip(afterRows[i]) {
			t.Fatalf("row %d text changed: %q -> %q", i, ansi.Strip(beforeRows[i]), ansi.Strip(afterRows[i]))
		}
		if lipgloss.Width(afterRows[i]) != lipgloss.Width(beforeRows[i]) {
			t.Fatalf("row %d width changed", i)
		}
	}
}

func TestMultiLineSelectionTrimsPadding(t *testing.T) {
	m := DefaultLiveChatModel()
	m.cfg.Raw = true
	m.messages = []LiveMessage{
		{Role: LiveRoleNotice, Content: "first line"},
		{Role: LiveRoleNotice, Content: "second line"},
	}

	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = model.(LiveChatModel)
	m.viewport.GotoTop()

	x0, y0 := screenOf(t, m, "line")
	x1, y1 := screenOf(t, m, "second")
	m = drag(m, x0, y0, x1+5, y1)

	got := m.selectedText()
	if !strings.HasPrefix(got, "line\n") || !strings.HasSuffix(got, "  second") {
		t.Fatalf("unexpected multi-line selection %q", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.HasSuffix(line, " ") {
			t.Fatalf("line kept its padding: %q", line)
		}
	}
}

func TestKeyPressClearsSelection(t *testing.T) {
	m := selectionChat(t)
	x, y := screenOf(t, m, "world")
	m = drag(m, x, y, x+2, y)

	var model tea.Model = m
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if model.(LiveChatModel).selection.active {
		t.Fatal("esc should clear the selection")
	}

	// ctrl+c with a selection copies rather than quitting.
	m = drag(model.(LiveChatModel), x, y, x+2, y)
	model, cmd := tea.Model(m).Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if model.(LiveChatModel).selection.active {
		t.Fatal("ctrl+c should clear the selection")
	}
	if cmd == nil {
		t.Fatal("ctrl+c with a selection should copy")
	}
}
