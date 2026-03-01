package ask

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/connordoman/doman/internal/txt"
)

func AskingSpinner(prompt string, actionWithError func(ctx context.Context) error) *spinner.Spinner {
	return spinner.New().Title(prompt).Style(lipgloss.NewStyle().Foreground(lipgloss.Color("#2563eb"))).ActionWithErr(actionWithError)
}

func FormatPrompt(prompt string, quick bool) string {
	terminalWidth := pkg.DetectTerminalWidth()
	promptStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Width(terminalWidth - 2)
	msg := fmt.Sprintf("%s %s", txt.Boldf("%s", txt.Bluef("You:")), prompt)
	if quick {
		msg = "⚡️ " + msg
	}
	return promptStyle.Render(msg)
}

func RandomSplashText() string {
	return SplashTexts[rand.Intn(len(SplashTexts))]
}
