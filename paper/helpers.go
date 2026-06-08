package paper

import "strings"

func stringsUpperTrim(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func stringsEqualFold(a string, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}
