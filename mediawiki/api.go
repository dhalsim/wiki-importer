package main

import "strings"

func apiBase(host string) string {
	if strings.HasSuffix(host, ".fandom.com") || host == "orthodoxwiki.org" {
		return "https://" + host + "/api.php"
	}

	return "https://" + host + "/w/api.php"
}
