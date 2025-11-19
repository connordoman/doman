package txt

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"
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

// countGraphemeClusters counts the number of user-perceived characters (grapheme clusters).
// This handles emojis, combining characters, and multi-rune sequences properly.
func countGraphemeClusters(s string) int {
	gr := uniseg.NewGraphemes(s)
	count := 0
	for gr.Next() {
		count++
	}
	return count
}

// isEmoji checks if a rune is likely an emoji.
func isEmoji(r rune) bool {
	// Check common emoji ranges
	return (r >= 0x1F300 && r <= 0x1F9FF) || // Misc Symbols and Pictographs, Emoticons, etc.
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
		(r >= 0x1F900 && r <= 0x1F9FF) // Supplemental Symbols and Pictographs
}

// containsEmoji checks if the string contains any emoji characters.
func containsEmoji(s string) bool {
	for _, r := range s {
		if isEmoji(r) {
			return true
		}
	}
	return false
}

// getDisplayWidth calculates the display width of a string, treating emojis consistently.
// For strings containing emojis, we count grapheme clusters and treat each emoji as width 2.
func getDisplayWidth(s string) int {
	// If the string contains emojis, use a simpler approach:
	// count grapheme clusters and assume each emoji takes 2 cells
	if containsEmoji(s) {
		graphemeCount := countGraphemeClusters(s)
		// For emoji-only strings or strings with emojis, assume width = grapheme count * 2
		// This is a simplification but works better for consistent alignment
		return graphemeCount * 2
	}

	// For non-emoji strings, check if it contains ANSI codes
	if strings.Contains(s, "\x1b[") {
		return lipgloss.Width(s)
	}

	// For simple ASCII strings, just use the rune count
	return len([]rune(s))
}

// PadRightWidth pads s on the right with spaces so that its display width is at least target.
// Handles emojis, ANSI escape codes, and wide characters properly.
func PadRightWidth(s string, target int) string {
	// For ANSI-colored strings (which don't contain emoji), use lipgloss
	if strings.Contains(s, "\x1b[") && !containsEmoji(s) {
		w := lipgloss.Width(s)
		if w >= target {
			return s
		}
		return s + strings.Repeat(" ", target-w)
	}

	// For emoji strings, use our custom width calculation
	w := getDisplayWidth(s)
	if w >= target {
		return s
	}
	return s + strings.Repeat(" ", target-w)
}

// PadLeftWidth pads s on the left with spaces so that its display width is at least target.
func PadLeftWidth(s string, target int) string {
	// For ANSI-colored strings (which don't contain emoji), use lipgloss
	if strings.Contains(s, "\x1b[") && !containsEmoji(s) {
		w := lipgloss.Width(s)
		if w >= target {
			return s
		}
		return strings.Repeat(" ", target-w) + s
	}

	// For emoji strings, use our custom width calculation
	w := getDisplayWidth(s)
	if w >= target {
		return s
	}
	return strings.Repeat(" ", target-w) + s
}
