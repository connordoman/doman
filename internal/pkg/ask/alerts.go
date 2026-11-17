package ask

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// alertKind is one of GitHub's alert types: > [!NOTE], > [!TIP], and so on.
type alertKind struct {
	title string
	dark  string
	light string
}

// Colours match GitHub's Primer palette for each alert in its dark and light
// themes.
var alertKinds = map[string]alertKind{
	"NOTE":      {title: "Note", dark: "#4493f8", light: "#0969da"},
	"TIP":       {title: "Tip", dark: "#3fb950", light: "#1a7f37"},
	"IMPORTANT": {title: "Important", dark: "#ab7df8", light: "#8250df"},
	"WARNING":   {title: "Warning", dark: "#d29922", light: "#9a6700"},
	"CAUTION":   {title: "Caution", dark: "#f85149", light: "#d1242f"},
}

var (
	// GitHub only treats a blockquote as an alert when the marker sits alone
	// on its first line.
	alertStartPattern = regexp.MustCompile(`(?i)^ {0,3}> ?\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$`)
	quoteLinePattern  = regexp.MustCompile(`^ {0,3}> ?`)
	fencePattern      = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
)

// markdownSegment is a run of ordinary Markdown, or the body of one alert
// when kind is set.
type markdownSegment struct {
	kind string
	body string
}

// splitAlerts separates top-level GitHub alerts from the Markdown around
// them. Alert markers inside fenced code blocks are left alone.
func splitAlerts(content string) []markdownSegment {
	lines := strings.Split(content, "\n")

	var (
		segments []markdownSegment
		plain    []string
		fence    string
	)

	flush := func() {
		if len(plain) > 0 {
			segments = append(segments, markdownSegment{body: strings.Join(plain, "\n")})
			plain = nil
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if m := fencePattern.FindStringSubmatch(line); m != nil {
			marker := m[1]
			switch {
			case fence == "":
				fence = marker
			case marker[0] == fence[0] && len(marker) >= len(fence) && strings.TrimSpace(line[strings.Index(line, marker)+len(marker):]) == "":
				fence = ""
			}
			plain = append(plain, line)
			continue
		}

		// A marker that follows another quoted line belongs to that quote.
		continuesQuote := i > 0 && quoteLinePattern.MatchString(lines[i-1])

		m := alertStartPattern.FindStringSubmatch(line)
		if fence != "" || m == nil || continuesQuote {
			plain = append(plain, line)
			continue
		}

		var body []string
		for i+1 < len(lines) && quoteLinePattern.MatchString(lines[i+1]) {
			i++
			body = append(body, quoteLinePattern.ReplaceAllString(lines[i], ""))
		}

		flush()
		segments = append(segments, markdownSegment{
			kind: strings.ToUpper(m[1]),
			body: strings.Join(body, "\n"),
		})
	}

	flush()
	return segments
}

// hasAlerts is a cheap check that lets plain Markdown skip the splitting
// pass entirely.
func hasAlerts(content string) bool {
	return strings.Contains(content, "[!")
}

// renderAlertBox draws a rounded box of the given outer width with the alert
// title set into its top border, fieldset-style. body is already rendered
// and wrapped to fit inside the box.
func renderAlertBox(kind alertKind, body string, width int, light bool) string {
	hex := kind.dark
	if light {
		hex = kind.light
	}
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
	title := accent.Bold(true).Render(kind.title)

	border := lipgloss.RoundedBorder()
	inner := max(width-2, 1)

	// ╭─ Title ─────╮
	titleWidth := ansi.StringWidth(kind.title)
	topFill := max(inner-titleWidth-3, 0)
	top := accent.Render(border.TopLeft+border.Top+" ") +
		title +
		accent.Render(" "+strings.Repeat(border.Top, topFill)+border.TopRight)

	contentWidth := max(inner-2, 1)
	rows := []string{top}
	if body != "" {
		for _, line := range strings.Split(body, "\n") {
			line = ansi.Truncate(line, contentWidth, "")
			pad := strings.Repeat(" ", max(contentWidth-ansi.StringWidth(line), 0))
			rows = append(rows, accent.Render(border.Left)+" "+line+pad+" "+accent.Render(border.Right))
		}
	}
	rows = append(rows, accent.Render(border.BottomLeft+strings.Repeat(border.Bottom, inner)+border.BottomRight))

	return strings.Join(rows, "\n")
}

// trimBlankLines drops leading and trailing lines that render as nothing but
// whitespace. Glamour pads every line to the wrap width, so a plain
// strings.Trim would not catch them.
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	isBlank := func(line string) bool { return strings.TrimSpace(ansi.Strip(line)) == "" }

	start, end := 0, len(lines)
	for start < end && isBlank(lines[start]) {
		start++
	}
	for end > start && isBlank(lines[end-1]) {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}
