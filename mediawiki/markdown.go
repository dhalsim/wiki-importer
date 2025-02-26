package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"fiatjaf/wiki-importer/common"
)

type PageResult struct {
	Parse struct {
		Title    string `json:"title"`
		Wikitext struct {
			All string `json:"*"`
		} `json:"wikitext"`
	} `json:"parse"`
}

func markdown(host string, page string) (string, string, error) {
	qs := url.Values{
		"action": {"parse"},
		"format": {"json"},
		"prop":   {"wikitext"},
		"page":   {page},
	}

	r, err := common.HttpGet(apiBase(host) + "?" + qs.Encode())
	if err != nil {
		return "", "", err
	}

	var res PageResult
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		r.Body.Close()
		return "", "", err
	}
	r.Body.Close()

	wikitext := strings.Builder{}
	wikitext.Grow(len(res.Parse.Wikitext.All))
	for _, line := range strings.Split(res.Parse.Wikitext.All, "\n") {
		if strings.HasPrefix(line, "{|") || strings.HasPrefix(line, "|") || strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "{{Clear}}") {
			continue
		}
		wikitext.WriteString(line)
	}

	cmd := exec.Command("pandoc", "--lua-filter", "mediawiki/wikilink.lua", "-f", "mediawiki", "-t", "commonmark", "-")
	cmd.Stdin = bytes.NewBufferString(wikitext.String())
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	markdown, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("pandoc error %w: %s", err, stderr.String())
	}

	return res.Parse.Title, string(markdown), nil
}
