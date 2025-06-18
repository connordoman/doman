package example

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var blue = lipgloss.Color("#3b82f6")

var (
	ExampleStyle = lipgloss.NewStyle().
			Bold(false).
			Foreground(blue).
			Background(lipgloss.AdaptiveColor{Light: "#f0f9ff", Dark: "#1e3a8a"}).
			Padding(0, 1).
			Margin(1, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(blue).
			BorderBackground(lipgloss.AdaptiveColor{Light: "#f0f9ff", Dark: "#1e3a8a"})

	ExampleCommand = &cobra.Command{
		Use:   "example",
		Short: "An example command to demonstrate functionality",
		Long:  "This command serves as an example to demonstrate how to create a command with Cobra and style it using Lipgloss.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(ExampleStyle.Render("Dominion Manager"))
			pkg.PrintInfo(pkg.Version())

		},
	}
)
