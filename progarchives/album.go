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

	fmt.Println("Fetching album from ", requestUrl)

	r, err := makeRequest(requestUrl)
	if err != nil {
		return "", "", err
	}

	ct := r.Header.Get("Content-Type")
	bodyReader, err := charset.NewReader(r.Body, ct)
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return "", "", err
	}
	r.Body.Close()

	title, err := getTitle(doc)
	if err != nil {
		return "", "", err
	}

	title = title + " (album)"

	artist := doc.Find(`h2`).Eq(0).Text()

	image, _ := doc.Find(`#imgCover`).Attr("src")
	imageUrl := "https://www.progarchives.com/" + image

	textContainer := doc.Find(`td`).Eq(1)

	text := strings.Builder{}
	for _, node := range textContainer.Children().Nodes {
		nodeToString(&text, node)
	}

	return title, fmt.Sprintf(`= %s

album from [[%s]]

image::%s[]

%s`,
		title, artist, imageUrl, text.String(),
	), nil
}
