package werpipe

import "strings"

func Normalize(text string) string {
	text = strings.TrimRight(text, "\n\r ")
	var result strings.Builder
	for _, r := range strings.ToLower(text) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '\'' {
			result.WriteRune(r)
		}
	}
	words := strings.Fields(result.String())
	return strings.Join(words, " ")
}
