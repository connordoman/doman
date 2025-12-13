package txt

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	italicStyle    = lipgloss.NewStyle().Italic(true)
	underlineStyle = lipgloss.NewStyle().Underline(true)
)

func Italicf(format string, args ...any) string {
	return italicStyle.Render(fmt.Sprintf(format, args...))
}

func Underlinef(format string, args ...any) string {
	return underlineStyle.Render(fmt.Sprintf(format, args...))
}
