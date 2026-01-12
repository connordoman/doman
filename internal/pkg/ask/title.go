package ask

import (
	"strings"
)

func FallbackTitleFromPrompt(prompt string) string {
	clean := strings.TrimSpace(prompt)
	if clean == "" {
		return "Untitled conversation"
	}

	firstLine := strings.SplitN(clean, "\n", 2)[0]
	return truncateTitle(firstLine, maxConversationTitleLength)
}

func IsMeaningfulTitle(title string) bool {
	if strings.TrimSpace(title) == "" {
		return false
	}

	return len([]rune(strings.TrimSpace(title))) >= minMeaningfulTitleRunes
}

func truncateTitle(title string, maxLen int) string {
	title = strings.TrimSpace(title)
	if maxLen <= 0 {
		return title
	}

	runes := []rune(title)
	if len(runes) <= maxLen {
		return title
	}

	return string(runes[:maxLen])
}
