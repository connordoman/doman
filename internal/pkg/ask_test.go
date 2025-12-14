package pkg

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFallbackTitleFromPrompt(t *testing.T) {
	title := FallbackTitleFromPrompt("Explain blue-green deployments\nwith examples")
	if title != "Explain blue-green deployments" {
		t.Fatalf("expected first line title, got %q", title)
	}
}

func TestTruncateTitleLimitsLength(t *testing.T) {
	long := strings.Repeat("x", maxConversationTitleLength+25)
	truncated := truncateTitle(long, maxConversationTitleLength)

	if utf8.RuneCountInString(truncated) != maxConversationTitleLength {
		t.Fatalf("expected length %d, got %d", maxConversationTitleLength, utf8.RuneCountInString(truncated))
	}
}

func TestIsMeaningfulTitle(t *testing.T) {
	if IsMeaningfulTitle("  ") {
		t.Fatal("whitespace title should not be meaningful")
	}

	if !IsMeaningfulTitle("API limits") {
		t.Fatal("non-empty title should be meaningful")
	}
}
