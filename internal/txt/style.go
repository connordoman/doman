package txt

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	italicStyle = lipgloss.NewStyle().Italic(true)
)

func Italicf(format string, args ...any) string {
	return italicStyle.Render(fmt.Sprintf(format, args...))
}
