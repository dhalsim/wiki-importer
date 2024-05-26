package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
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

	r, err := http.Get("https://" + host + "/w/api.php?" + qs.Encode())
	if err != nil {
		return "", "", err
	}

	var res PageResult
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		r.Body.Close()
		return "", "", err
	}
	r.Body.Close()

	cmd := exec.Command("pandoc", "--lua-filter", "mediawiki/wikilink.lua", "-f", "mediawiki", "-t", "commonmark", "-")
	cmd.Stdin = bytes.NewBufferString(res.Parse.Wikitext.All)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	markdown, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("pandoc error %w: %s", err, stderr.String())
	}

	return res.Parse.Title, string(markdown), nil
}
