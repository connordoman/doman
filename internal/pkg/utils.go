package pkg

import (
	"fmt"
	"os"

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

func PrintSuccess(format string, args ...any) {
	str := fmt.Sprintf("✔ "+format, args...)
	fmt.Println(successStyle.Render(str))
}

func PrintError(format string, args ...any) {
	str := fmt.Sprintf("✘ "+format, args...)
	fmt.Println(errorStyle.Render(str))
}

func PrintInfo(format string, args ...any) {
	fmt.Println(greyStyle.Render(fmt.Sprintf(format, args...)))
}

func FailAndExit(format string, args ...any) {
	PrintError(format, args...)
	os.Exit(1)
}

func SprintfBold(format string, args ...any) string {
	return boldStyle.Render(fmt.Sprintf(format, args...))
}

func SprintfGrey(format string, args ...any) string {
	return greyStyle.Render(fmt.Sprintf(format, args...))
}
