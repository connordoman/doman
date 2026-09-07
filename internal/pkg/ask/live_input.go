package ask

import (
	"regexp"
	"strings"
)

// QuickCommand is the inline command that switches a single turn into quick
// mode. It may appear anywhere in the message.
const QuickCommand = "/quick"

// quickCommandPattern matches a standalone '/quick' token, preserving the
// whitespace in front of it so line breaks in the message survive stripping.
var quickCommandPattern = regexp.MustCompile(`(?i)(^|\s)` + regexp.QuoteMeta(QuickCommand) + `(?:\s|$)`)

// trailingSpacePattern matches spaces and tabs left at the end of a line once
// the command has been removed.
var trailingSpacePattern = regexp.MustCompile(`[ \t]+(\n|$)`)

// ParseQuickCommand strips every occurrence of '/quick' from input and reports
// whether at least one was found. The remaining text keeps its original line
// structure so multi-line prompts are not mangled.
func ParseQuickCommand(input string) (prompt string, quick bool) {
	prompt = input

	for {
		stripped := quickCommandPattern.ReplaceAllString(prompt, "$1")
		if stripped == prompt {
			break
		}
		prompt = stripped
		quick = true
	}

	prompt = trailingSpacePattern.ReplaceAllString(prompt, "$1")

	return strings.TrimSpace(prompt), quick
}
