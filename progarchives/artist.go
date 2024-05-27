package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func artist(id int) (string, string, error) {
	req, err := http.NewRequest("GET", "https://www.progarchives.com/artist.asp?"+url.Values{"id": {strconv.Itoa(id)}}.Encode(), nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("user-agent", "Chrome")
	req.Header.Add("accept", "text/html")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return "", "", err
	}
	r.Body.Close()

	title := removeNonUtf8(doc.Find(`h1`).Text())
	if title == "" {
		return "", "", fmt.Errorf("title error")
	}

	cat := doc.Find(`h2`).First().Text()
	if cat == "" || !strings.Contains(cat, "•") {
		return "", "", fmt.Errorf("category error: %s", cat)
	}

	spl := strings.Split(cat, "•")
	category := strings.TrimSpace(spl[0])
	country := strings.TrimSpace(spl[1])

	image := fmt.Sprintf("https://www.progarchives.com/progressive_rock_discography_band/%d.png", id)

	bioStart := doc.Find("strong").Eq(1)
	bioContainer := bioStart.Parent()
	bioStart.Remove()
	var bio *goquery.Selection
	moreBio := bioContainer.Find("#moreBio")
	if moreBio.Length() == 0 {
		bio = bioContainer
	} else {
		bio = moreBio
	}

	bioText := strings.Builder{}
	for _, node := range bio.Nodes {
		nodeToString(&bioText, node)
	}

	discography := strings.Builder{}
	doc.Find("#discography").NextUntil("table").Next().Find("td").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a > strong").Text()
		year := s.Find("a + br + span").Text()
		discography.WriteString("\n  - [[")
		discography.WriteString(removeNonUtf8(title))
		discography.WriteString("]] ([[")
		discography.WriteString(removeNonUtf8(year))
		discography.WriteString("]])")
	})

	return title, fmt.Sprintf("# %s\n\n[[%s]], [[%s]]\n\n![](%s)\n\n%s\n\n## Discography\n%s",
		title, category, country, image, bioText.String(), discography.String(),
	), nil
}
