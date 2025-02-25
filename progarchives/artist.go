package progarchives

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

func artist(id uint64) (string, string, error) {
	params := url.Values{"id": {strconv.FormatUint(id, 10)}}
	requestUrl := "https://www.progarchives.com/artist.asp?" + params.Encode()

	logger.Printf("Fetching artist from %s\n", requestUrl)

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

	logger.Printf("Processing artist: %s\n", title)

	cat := doc.Find(`h2`).First().Text()
	if cat == "" || !strings.Contains(cat, "•") {
		return "", "", fmt.Errorf("category error: %s", cat)
	}

	spl := strings.Split(cat, "•")
	category := strings.TrimSpace(spl[0])
	country := strings.TrimSpace(spl[1])

	image, imageFound := doc.Find(`meta[property="og:image"]`).Attr("content")

	if !imageFound {
		image = fmt.Sprintf("https://www.progarchives.com/progressive_rock_discography_band/%d.jpg", id)
	}

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
		albumTitle := s.Find("a > strong").Text()
		albumYear := s.Find("a + br + span").Text()

		discography.WriteString(fmt.Sprintf(`
- [[%s (album)]] ([[%s]])
`, albumTitle, albumYear))
	})

	return title, fmt.Sprintf(`[[%s]], [[%s]]

image::%s[]

%s

== Discography

%s`,
		category, country, image, bioText.String(), discography.String(),
	), nil
}
