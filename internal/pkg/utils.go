package pkg

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var successStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#22c55e"))

var errorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ef4444"))

func PrintSuccess(format string, args ...any) {
	str := fmt.Sprintf("✔ "+format, args...)
	fmt.Println(successStyle.Render(str))
}

func PrintError(format string, args ...any) {
	str := fmt.Sprintf("✘ "+format, args...)
	fmt.Println(errorStyle.Render(str))
}

func FailAndExit(format string, args ...any) {
	PrintError(format, args...)
	os.Exit(1)
}
