package main

import "strings"

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
