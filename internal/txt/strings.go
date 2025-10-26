package txt

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Repeat(char rune, length int) string {
	return strings.Repeat(string(char), length)

}

func Line(length int) string {
	return Repeat('─', length)
}

func Separator() string {
	return Line(40)
}

// PadRightWidth pads s on the right with spaces so that its display width is at least target.
// Uses wcwidth via go-runewidth to properly handle emoji and wide runes.
func PadRightWidth(s string, target int) string {
	w := lipgloss.Width(s)
	if w >= target {
		return s
	}
	return s + strings.Repeat(" ", target-w)
}

// PadLeftWidth pads s on the left with spaces so that its display width is at least target.
func PadLeftWidth(s string, target int) string {
	w := lipgloss.Width(s)
	if w >= target {
		return s
	}
	return strings.Repeat(" ", target-w) + s
}
