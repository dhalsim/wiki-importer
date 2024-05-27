package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func album(id int) (string, string, error) {
	req, err := http.NewRequest("GET", "https://www.progarchives.com/album.asp?"+url.Values{"id": {strconv.Itoa(id)}}.Encode(), nil)
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

	artist := removeNonUtf8(doc.Find(`h2`).Eq(0).Text())

	image, _ := doc.Find(`#imgCover`).Attr("src")

	textContainer := doc.Find(`td`).Eq(1)

	text := strings.Builder{}
	for _, node := range textContainer.Children().Nodes {
		nodeToString(&text, node)
	}

	return title, fmt.Sprintf("# %s\n\nalbum from [[%s]]\n\n![](%s)\n\n%s",
		title, artist, image, text.String(),
	), nil
}
