package ask

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestSplitAlerts(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []markdownSegment
	}{
		{
			name: "plain markdown",
			in:   "Hello\n\n> a quote",
			want: []markdownSegment{{body: "Hello\n\n> a quote"}},
		},
		{
			name: "alert between paragraphs",
			in:   "Before\n\n> [!NOTE]\n> First line\n>\n> Second\n\nAfter",
			want: []markdownSegment{
				{body: "Before\n"},
				{kind: "NOTE", body: "First line\n\nSecond"},
				{body: "\nAfter"},
			},
		},
		{
			name: "case insensitive marker",
			in:   "> [!warning]\n> Careful",
			want: []markdownSegment{{kind: "WARNING", body: "Careful"}},
		},
		{
			name: "marker with trailing text is a plain quote",
			in:   "> [!NOTE] not alone\n> body",
			want: []markdownSegment{{body: "> [!NOTE] not alone\n> body"}},
		},
		{
			name: "unknown kind is a plain quote",
			in:   "> [!HINT]\n> body",
			want: []markdownSegment{{body: "> [!HINT]\n> body"}},
		},
		{
			name: "marker inside an existing quote",
			in:   "> intro\n> [!NOTE]\n> body",
			want: []markdownSegment{{body: "> intro\n> [!NOTE]\n> body"}},
		},
		{
			name: "marker inside a code fence",
			in:   "```md\n> [!NOTE]\n> body\n```",
			want: []markdownSegment{{body: "```md\n> [!NOTE]\n> body\n```"}},
		},
		{
			name: "alert after a closed fence",
			in:   "~~~\ncode\n~~~\n> [!TIP]\n> body",
			want: []markdownSegment{
				{body: "~~~\ncode\n~~~"},
				{kind: "TIP", body: "body"},
			},
		},
		{
			name: "marker only, body still streaming",
			in:   "> [!CAUTION]",
			want: []markdownSegment{{kind: "CAUTION", body: ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAlerts(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d segments %q, want %d %q", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("segment %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRenderAlertBox(t *testing.T) {
	box := renderAlertBox(alertKinds["IMPORTANT"], "hello", 30, false)
	lines := strings.Split(ansi.Strip(box), "\n")

	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3:\n%s", len(lines), ansi.Strip(box))
	}
	if !strings.HasPrefix(lines[0], "╭─ Important ─") || !strings.HasSuffix(lines[0], "╮") {
		t.Errorf("top border = %q", lines[0])
	}
	if lines[1] != "│ hello"+strings.Repeat(" ", 21)+" │" {
		t.Errorf("body = %q", lines[1])
	}
	for i, line := range lines {
		if w := ansi.StringWidth(line); w != 30 {
			t.Errorf("line %d width = %d, want 30: %q", i, w, line)
		}
	}
}

func TestMarkdownRendererAlerts(t *testing.T) {
	r, err := NewMarkdownRenderer(50)
	if err != nil {
		t.Fatal(err)
	}

	out, err := r.Render("Intro\n\n> [!TIP]\n> Use **bold** text.\n\nOutro")
	if err != nil {
		t.Fatal(err)
	}
	plain := ansi.Strip(out)

	for _, want := range []string{"Intro", "╭─ Tip ─", "Use bold text.", "Outro"} {
		if !strings.Contains(plain, want) {
			t.Errorf("output missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "[!TIP]") {
		t.Errorf("alert marker leaked into output:\n%s", plain)
	}
	for _, line := range strings.Split(plain, "\n") {
		if w := ansi.StringWidth(line); w > 50 {
			t.Errorf("line wider than wrap width (%d): %q", w, line)
		}
	}
}
