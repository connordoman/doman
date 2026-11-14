package ask

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/connordoman/windy"

	"doman.sh/doman/internal/pkg"
)

const (
	// liveTranscriptLeft is the screen column where the transcript starts,
	// after the one column gutter.
	liveTranscriptLeft = 1

	liveAutoScrollInterval = 50 * time.Millisecond
	liveFlashDuration      = 2 * time.Second
)

// selectionPoint addresses a single cell in the transcript. line is an index
// into the full transcript, not the visible window, so a selection stays put
// while the viewport scrolls underneath it.
type selectionPoint struct {
	line int
	col  int
}

func (p selectionPoint) before(o selectionPoint) bool {
	return p.line < o.line || (p.line == o.line && p.col < o.col)
}

// liveSelection is a mouse selection over the transcript. Both ends are
// inclusive, so the cell that was clicked to start the drag is always part of
// the selection.
type liveSelection struct {
	active   bool
	dragging bool
	moved    bool
	anchor   selectionPoint
	cursor   selectionPoint

	// lastY is the most recent pointer row, used to keep auto-scrolling while
	// the pointer rests above or below the transcript.
	lastY      int
	autoScroll bool
	lastX      int
}

// bounds returns the selection's ends in reading order.
func (s liveSelection) bounds() (start, end selectionPoint) {
	if s.cursor.before(s.anchor) {
		return s.cursor, s.anchor
	}
	return s.anchor, s.cursor
}

// span reports the half-open cell range [from, to) selected on line, and
// whether the line is part of the selection at all. to is -1 when the
// selection runs to the end of the line.
func (s liveSelection) span(line int) (from, to int, ok bool) {
	if !s.active {
		return 0, 0, false
	}

	start, end := s.bounds()
	if line < start.line || line > end.line {
		return 0, 0, false
	}

	from, to = 0, -1
	if line == start.line {
		from = start.col
	}
	if line == end.line {
		to = end.col + 1
	}

	return from, to, true
}

// copyToClipboard writes to the native clipboard. Tests replace it so they do
// not clobber the clipboard of whoever runs them.
var copyToClipboard = pkg.CopyToClipboard

type liveAutoScrollMsg struct{}

type liveFlashExpiredMsg struct{ id int }

// handleMouse drives selection. It reports whether the event was consumed so
// that the viewport does not also react to it.
func (m *LiveChatModel) handleMouse(msg tea.MouseMsg) (tea.Cmd, bool) {
	mouse := msg.Mouse()

	switch msg.(type) {
	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft {
			return nil, false
		}

		if !m.inTranscript(mouse.X, mouse.Y) {
			m.selection = liveSelection{}
			return nil, true
		}

		p := m.pointAt(mouse.X, mouse.Y)
		m.selection = liveSelection{
			active:   true,
			dragging: true,
			anchor:   p,
			cursor:   p,
			lastX:    mouse.X,
			lastY:    mouse.Y,
		}
		return nil, true

	case tea.MouseMotionMsg:
		if !m.selection.dragging {
			return nil, false
		}

		m.selection.lastX = mouse.X
		m.selection.lastY = mouse.Y
		m.scrollTowardPointer()
		m.selection.cursor = m.pointAt(mouse.X, mouse.Y)
		m.selection.moved = true

		if m.outsideTranscriptRows(mouse.Y) && !m.selection.autoScroll {
			m.selection.autoScroll = true
			return autoScrollTick(), true
		}
		return nil, true

	case tea.MouseReleaseMsg:
		if !m.selection.dragging {
			return nil, false
		}

		m.selection.dragging = false
		m.selection.autoScroll = false

		// A plain click with no drag clears rather than selecting one cell.
		if !m.selection.moved {
			m.selection = liveSelection{}
			return nil, true
		}

		return m.copySelection(), true
	}

	return nil, false
}

// handleAutoScroll keeps the transcript moving while a drag rests past its top
// or bottom edge, extending the selection as new lines come into view.
func (m *LiveChatModel) handleAutoScroll() tea.Cmd {
	if !m.selection.dragging || !m.outsideTranscriptRows(m.selection.lastY) {
		m.selection.autoScroll = false
		return nil
	}

	m.scrollTowardPointer()
	m.selection.cursor = m.pointAt(m.selection.lastX, m.selection.lastY)
	return autoScrollTick()
}

func autoScrollTick() tea.Cmd {
	return tea.Tick(liveAutoScrollInterval, func(time.Time) tea.Msg { return liveAutoScrollMsg{} })
}

func (m *LiveChatModel) scrollTowardPointer() {
	top := liveHeaderHeight
	bottom := top + m.viewport.Height()

	switch {
	case m.selection.lastY < top:
		m.viewport.ScrollUp(1)
	case m.selection.lastY >= bottom:
		m.viewport.ScrollDown(1)
	}
}

// copySelection writes the selected text to the clipboard and flashes a
// confirmation in the status line.
func (m *LiveChatModel) copySelection() tea.Cmd {
	text := m.selectedText()
	if text == "" {
		return nil
	}

	// The native clipboard covers local sessions; OSC52 covers terminals
	// reached over SSH, where the native clipboard belongs to the wrong host.
	cmds := []tea.Cmd{tea.SetClipboard(text)}
	if err := copyToClipboard(text); err != nil {
		cmds = append(cmds, m.flash("Copied via terminal clipboard"))
	} else {
		n := utf8.RuneCountInString(text)
		newLines := strings.Count(text, "\n")
		noun := "characters"
		if newLines > 1 {
			noun = "lines"
			n = newLines
		} else if n == 1 {
			noun = "character"
		}

		cmds = append(cmds, m.flash(fmt.Sprintf("Copied %d %s", n, noun)))
	}

	return tea.Batch(cmds...)
}

func (m *LiveChatModel) flash(text string) tea.Cmd {
	m.flashID++
	m.flashText = text
	id := m.flashID
	return tea.Tick(liveFlashDuration, func(time.Time) tea.Msg { return liveFlashExpiredMsg{id: id} })
}

// selectedText returns the plain text under the selection. Rendered lines are
// padded to the transcript width, so trailing blanks are trimmed per line.
func (m LiveChatModel) selectedText() string {
	if !m.selection.active || len(m.transcriptLines) == 0 {
		return ""
	}

	start, end := m.selection.bounds()
	end.line = min(end.line, len(m.transcriptLines)-1)

	lines := make([]string, 0, end.line-start.line+1)
	for i := start.line; i <= end.line; i++ {
		plain := ansi.Strip(m.transcriptLines[i])
		from, to, _ := m.selection.span(i)
		if to < 0 {
			to = ansi.StringWidth(plain)
		}
		lines = append(lines, strings.TrimRight(ansi.Cut(plain, from, to), " "))
	}

	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// highlightSelection paints the selection over the rendered viewport. view is
// the viewport's output, whose first row is transcript line offset.
func (m LiveChatModel) highlightSelection(view string, offset int) string {
	if !m.selection.active {
		return view
	}

	style := lipgloss.NewStyle().
		Background(liveColor(windy.Blue800)).
		Foreground(liveColor(windy.Neutral100))

	rows := strings.Split(view, "\n")
	for i, row := range rows {
		from, to, ok := m.selection.span(offset + i)
		if !ok {
			continue
		}

		// Stop at the end of the text rather than the padded edge of the
		// viewport, the way a terminal's own selection looks.
		textWidth := ansi.StringWidth(strings.TrimRight(ansi.Strip(row), " "))
		if to < 0 || to > textWidth {
			to = textWidth
		}

		rows[i] = highlightCells(row, from, to, style)
	}

	return strings.Join(rows, "\n")
}

// highlightCells restyles cells [from, to) of a styled line. The text after
// the selection keeps its own styling because ansi.Cut replays the escape
// sequences that precede the cut.
func highlightCells(line string, from, to int, style lipgloss.Style) string {
	width := ansi.StringWidth(line)
	to = min(to, width)
	if from >= to {
		return line
	}

	left := ansi.Cut(line, 0, from)
	middle := ansi.Strip(ansi.Cut(line, from, to))
	right := ansi.Cut(line, to, width)

	return left + ansi.ResetStyle + style.Render(middle) + right
}

func (m LiveChatModel) inTranscript(x, y int) bool {
	return x >= liveTranscriptLeft && !m.outsideTranscriptRows(y)
}

func (m LiveChatModel) outsideTranscriptRows(y int) bool {
	return y < liveHeaderHeight || y >= liveHeaderHeight+m.viewport.Height()
}

// pointAt maps a screen cell to a transcript cell, clamping to the visible
// window so a drag past the edges selects up to the edge.
func (m LiveChatModel) pointAt(x, y int) selectionPoint {
	row := min(max(y-liveHeaderHeight, 0), m.viewport.Height()-1)
	line := m.viewport.YOffset() + row
	line = min(max(line, 0), max(len(m.transcriptLines)-1, 0))

	return selectionPoint{
		line: line,
		col:  max(x-liveTranscriptLeft, 0),
	}
}
