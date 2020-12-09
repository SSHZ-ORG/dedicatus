package utils

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Normalizes Alias by converting to lower case and converting everything to NFKC form.
func NormalizeAlias(alias string) string {
	return strings.ToLower(norm.NFKC.String(alias))
}

func TrimFirstRune(s string) string {
	for i := range s {
		if i > 0 {
			return s[i:]
		}
	}
	return ""
}
