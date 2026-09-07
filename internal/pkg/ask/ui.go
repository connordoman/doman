package ask

import (
	"context"
	"fmt"
	"math/rand"

	"doman.sh/internal/pkg"
	"doman.sh/internal/txt"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/connordoman/windy"
)

func AskingSpinner(prompt string, quick bool, actionWithError func(ctx context.Context) error) *spinner.Spinner {
	color := windy.Blue500
	if quick {
		color = windy.Yellow500
	}
	return spinner.New().Title(prompt).Style(lipgloss.NewStyle().Foreground(color.Glossy())).ActionWithErr(actionWithError)
}

func FormatPrompt(prompt string, quick bool) string {
	terminalWidth := pkg.DetectTerminalWidth()
	promptStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Width(terminalWidth - 2).BorderForeground(windy.Neutral500.Glossy())
	msg := fmt.Sprintf("%s %s", txt.Boldf("%s", txt.Bluef("You:")), prompt)
	if quick {
		msg = "⚡️ " + msg
	}
	return promptStyle.Render(msg)
}

func RandomSplashText() string {
	return SplashTexts[rand.Intn(len(SplashTexts))]
}
