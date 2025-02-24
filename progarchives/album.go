package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func album(id int) (string, string, error) {
	params := url.Values{"id": {strconv.Itoa(id)}}
	requestUrl := "https://www.progarchives.com/album.asp?" + params.Encode()

	logger.Printf("Fetching album from %s\n", requestUrl)

	r, err := makeRequest(requestUrl)
	if err != nil {
		return "", "", fmt.Errorf("request failed: %w", err)
	}

	ct := r.Header.Get("Content-Type")
	bodyReader, err := charset.NewReader(r.Body, ct)
	if err != nil {
		return "", "", fmt.Errorf("charset reader failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return "", "", fmt.Errorf("goquery parse failed: %w", err)
	}
	r.Body.Close()

	title, err := getTitle(doc)
	if err != nil {
		return "", "", fmt.Errorf("get title failed: %w", err)
	}

	logger.Printf("Processing album: %s\n", title)

	title = title + " (album)"

	artist := doc.Find(`h2`).Eq(0).Text()

	image, _ := doc.Find(`#imgCover`).Attr("src")
	imageUrl := "https://www.progarchives.com/" + image

	textContainer := doc.Find(`td`).Eq(1)

	text := strings.Builder{}
	for _, node := range textContainer.Children().Nodes {
		nodeToString(&text, node)
	}

	return title, fmt.Sprintf(`album from [[%s]]

image::%s[]

%s`,
		artist, imageUrl, text.String(),
	), nil
}
