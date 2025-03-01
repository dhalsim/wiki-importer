package mediawiki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
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

func asciidoc(host string, pageTitle string) (string, string, error) {
	qs := url.Values{
		"action": {"parse"},
		"format": {"json"},
		"prop":   {"wikitext"},
		"page":   {pageTitle},
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

	asciidoc, err := parseWikitext(res, "mediawiki/lua")
	if err != nil {
		return "", "", err
	}

	return res.Parse.Title, asciidoc, nil
}

func parseWikitext(res PageResult, luaFolder string) (string, error) {
	wikitext := strings.Builder{}
	content := res.Parse.Wikitext.All

	// Do all replacements at once
	content = strings.NewReplacer(
		`\n`, "\n",
		`\u003C`, "<",
		`\u003E`, ">",
	).Replace(content)

	// Pre-allocate capacity
	wikitext.Grow(len(content))

	// Write filtered content
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "{|") || strings.HasPrefix(line, "|") || strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "{{Clear}}") {
			continue
		}
		wikitext.WriteString(line + "\n")
	}

	cmd := exec.Command(
		"pandoc",
		// filter order is important
		"--lua-filter", filepath.Join(luaFolder, "remove-header-ids.lua"),
		"--lua-filter", filepath.Join(luaFolder, "description-list.lua"),
		"--lua-filter", filepath.Join(luaFolder, "wikilink.lua"),
		"-f", "mediawiki",
		"-t", "asciidoc",
		"--wrap=none",
		"-",
	)

	cmd.Stdin = bytes.NewBufferString(wikitext.String())
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	asciidoc, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pandoc error %w: %s", err, stderr.String())
	}

	return string(asciidoc), nil
}
