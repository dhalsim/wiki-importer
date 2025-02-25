package progarchives

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func nodeToString(str *strings.Builder, node *html.Node) {
	switch node.Type {
	case html.TextNode:
		str.WriteString(removeNonUtf8(node.Data))
	case html.ElementNode:
		switch node.Data {
		case "span":
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
		case "br":
			str.WriteByte('\n')
		case "p":
			str.WriteByte('\n')
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
			str.WriteByte('\n')
		case "b":
			str.WriteString("**")
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
			str.WriteString("**")
		case "i":
			str.WriteByte('_')
			for nextNode := node.FirstChild; nextNode != nil; nextNode = nextNode.NextSibling {
				nodeToString(str, nextNode)
			}
			str.WriteByte('_')
		case "a":
			if node.FirstChild == nil {
				return
			}

			targetIdx := slices.IndexFunc(node.Attr, func(a html.Attribute) bool { return a.Key == "href" })
			if targetIdx == -1 {
				str.WriteString(removeNonUtf8(node.FirstChild.Data))
				return
			}

			target := node.Attr[targetIdx].Val
			if strings.HasPrefix(target, "http") || !strings.Contains(target, ".asp") {
				str.WriteString(fmt.Sprintf("%s[%s]", target, node.FirstChild.Data))
			} else {
				str.WriteString(fmt.Sprintf("[[%s]]", node.FirstChild.Data))
			}
		}
	}
}

func removeNonUtf8(str string) string {
	var sb strings.Builder
	for _, r := range str {
		if utf8.ValidRune(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func getHttpClient() *http.Client {
	transport := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
		},
	}
	return &http.Client{
		Transport: transport,
	}
}

func getTitle(doc *goquery.Document) (string, error) {
	title := doc.Find(`h1`).Text()

	if title == "" {
		return "", fmt.Errorf("title error")
	}

	return title, nil
}

func makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("user-agent", "Chrome")
	req.Header.Add("accept", "text/html")

	httpClient := getHttpClient()

	r, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", r.StatusCode)
	}

	return r, nil
}
