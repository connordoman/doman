package txt

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#22c55e"))

var errorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#ef4444"))

var boldStyle = lipgloss.NewStyle().
	Bold(true)

var greyStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#6b7280"))

var blueStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#3b82f6"))

func Boldf(format string, args ...any) string {
	return boldStyle.Render(fmt.Sprintf(format, args...))
}

func Greyf(format string, args ...any) string {
	return greyStyle.Render(fmt.Sprintf(format, args...))
}

func Bluef(format string, args ...any) string {
	return blueStyle.Render(fmt.Sprintf(format, args...))
}

func Successf(format string, args ...any) string {
	return successStyle.Render(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) string {
	return errorStyle.Render(fmt.Sprintf(format, args...))
}
