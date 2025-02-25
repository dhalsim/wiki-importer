package movies

import (
	"fmt"
	"strings"
)

func splitAndWikilink(s string) string {
	if s != "" {
		spl := strings.Split(s, ", ")
		for i, item := range spl {
			spl[i] = "[[" + item + "]]"
		}
		return strings.Join(spl, ", ")
	}
	return ""
}

func getYesterdays(format string) string {
	return fmt.Sprintf(
		format,
		yesterday.Month(),
		yesterday.Day(),
		yesterday.Year(),
	)
}
