package txt

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22c55e"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ef4444"))

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#f59e0b"))

	boldStyle = lipgloss.NewStyle().
			Bold(true)

	greyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280"))

	blueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6"))
)

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

func Warningf(format string, args ...any) string {
	return warningStyle.Render(fmt.Sprintf(format, args...))
}

// ----- Gradient color interpolation utilities -----

// RGB represents a color in 0..255 per channel.
type RGB struct{ R, G, B int }

// ColorStop is a color at a given position t in [0,1].
type ColorStop struct {
	T     float64
	Color RGB
}

// Gradient defines an ordered list of color stops.
type Gradient []ColorStop

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func lerp(a, b float64, t float64) float64 { return a + (b-a)*t }

func lerpRGB(a, b RGB, t float64) RGB {
	return RGB{
		R: int(lerp(float64(a.R), float64(b.R), t) + 0.5),
		G: int(lerp(float64(a.G), float64(b.G), t) + 0.5),
		B: int(lerp(float64(a.B), float64(b.B), t) + 0.5),
	}
}

func (g Gradient) ColorAt(t float64) RGB {
	if len(g) == 0 {
		return RGB{R: 255, G: 255, B: 255}
	}
	t = clamp01(t)
	// If t is before first or after last, clamp
	if t <= g[0].T {
		return g[0].Color
	}
	if t >= g[len(g)-1].T {
		return g[len(g)-1].Color
	}
	// Find the surrounding stops
	for i := 0; i < len(g)-1; i++ {
		a := g[i]
		b := g[i+1]
		if t >= a.T && t <= b.T {
			// normalize
			span := b.T - a.T
			if span <= 0 {
				return b.Color
			}
			local := (t - a.T) / span
			return lerpRGB(a.Color, b.Color, local)
		}
	}
	return g[len(g)-1].Color
}

func hex(c RGB) string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

// GreenYellowRedGradient returns a default gradient from green (0) to yellow (0.5) to red (1.0).
func GreenYellowRedGradient() Gradient {
	return Gradient{
		{T: 0.0, Color: RGB{R: 34, G: 197, B: 94}}, // #22c55e
		{T: 0.5, Color: RGB{R: 234, G: 179, B: 8}}, // #eab308
		{T: 1.0, Color: RGB{R: 239, G: 68, B: 68}}, // #ef4444
	}
}

func TemperatureGradient() Gradient {
	return Gradient{
		{T: 0.1, Color: RGB{R: 37, G: 99, B: 235}}, // #2563eb
		{T: 0.4, Color: RGB{R: 6, G: 182, B: 212}}, // #06b6d4
		{T: 0.6, Color: RGB{R: 22, G: 163, B: 74}}, // #16a34a
		{T: 0.7, Color: RGB{R: 22, G: 163, B: 74}}, // #16a34a
		// {T: 0.4, Color: RGB{R: 22, G: 163, B: 74}},  // #16a34a

		{T: 0.8, Color: RGB{R: 250, G: 204, B: 21}}, // #facc15
		{T: 1.0, Color: RGB{R: 239, G: 68, B: 68}},  // #ef4444
	}
}

// ColorizePercent renders text using a gradient color mapped from percent in [0,100].
func ColorizePercent(text string, percent float64, gradient Gradient) string {
	t := percent / 100.0
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	rgb := gradient.ColorAt(t)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(hex(rgb)))
	return style.Render(text)
}
