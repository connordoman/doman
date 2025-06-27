package txt

import "strings"

func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
