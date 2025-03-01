package names

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"fiatjaf/wiki-importer/common"

	"github.com/PuerkitoBio/goquery"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
	"golang.org/x/net/html"
)

// BehindTheNameParams handles Nostr configuration and operations for the behindthename package
type BehindTheNameParams struct {
	nostrKey     string
	pool         *nostr.SimplePool
	relay        string
	continueFrom int
	logger       *log.Logger
}

func NewBehindTheNameParams(
	key string,
	p *nostr.SimplePool,
	r string,
	continueFrom int,
	logger *log.Logger,
) *BehindTheNameParams {
	return &BehindTheNameParams{
		nostrKey:     key,
		pool:         p,
		relay:        r,
		continueFrom: continueFrom,
		logger:       logger,
	}
}

func HandleBehindthename(ctx context.Context, params *BehindTheNameParams) error {
	if params.continueFrom == 0 {
		params.continueFrom = 1
	}

	for i := params.continueFrom; i < 101; i++ {
		params.logger.Printf("Processing page %d", i)

		proceed, err := doPage(ctx, params, i)
		if err != nil {
			return fmt.Errorf("process page %d: %w", i, err)
		}

		if !proceed {
			break
		}

		time.Sleep(time.Second * 5)
	}

	return nil
}

func doPage(ctx context.Context, params *BehindTheNameParams, num int) (bool, error) {
	var resp *http.Response
	var err error

	for {
		resp, err = common.HttpGet(fmt.Sprintf("https://www.behindthename.com/names/%d", num))
		if err != nil {
			params.logger.Printf("error fetching page %d: %v, retrying in 5 minutes", num, err)

			time.Sleep(time.Minute * 5)

			continue
		}

		break
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, fmt.Errorf("parse page HTML: %w", err)
	}

	sel := doc.Find(`.listname a[href^="/name/"]`)
	if len(sel.Nodes) == 0 {
		return false, nil
	}

	sel.Each(func(_ int, sn *goquery.Selection) {
		href, _ := sn.Attr("href")
		href = strings.TrimSpace(href)
		name := strings.TrimSpace(sn.Text())

		if err := doName(ctx, params, "https://www.behindthename.com"+href, name); err != nil {
			params.logger.Printf("error processing name %s: %v", name, err)
		}
	})

	return true, nil
}

func doName(ctx context.Context, params *BehindTheNameParams, url string, name string) error {
	var resp *http.Response
	var err error

	for {
		resp, err = common.HttpGet(url)
		if err != nil {
			params.logger.Printf("error fetching name %s: %v, retrying in 5 minutes", name, err)

			time.Sleep(5 * time.Minute)

			continue
		}

		break
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("parse name page HTML: %w", err)
	}

	nameNodes := doc.Find(".namedef").Nodes
	if len(nameNodes) == 0 {
		return fmt.Errorf("no name definition found for %s", name)
	}

	def := ""
	for c := nameNodes[0].FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			def += c.Data
		case html.ElementNode:
			switch c.Data {
			case "em", "span", "small":
				if c.FirstChild != nil {
					def += c.FirstChild.Data
				}
			case "a":
				if c.FirstChild == nil {
					params.logger.Printf("Skipping empty link tag in definition for %s", name)
					continue
				}
				if slices.ContainsFunc(c.Attr, func(a html.Attribute) bool { return a.Key == "href" && strings.HasPrefix(a.Val, "/name/") }) {
					name := strings.TrimSpace(c.FirstChild.Data)
					if name == "" {
						params.logger.Printf("Skipping link with empty text in definition for %s", name)
						continue
					}
					def += "[[" + name + "]]"
				} else {
					def += c.FirstChild.Data
				}
			case "i":
				if c.FirstChild != nil {
					def += "_" + c.FirstChild.Data + "_"
				}
			case "b":
				if c.FirstChild != nil {
					def += "**" + c.FirstChild.Data + "**"
				}
			}
		}
	}
	def = strings.TrimSpace(def)
	def += "\n\nhttps://www.behindthename.com/name/" + strings.Split(url, "/name/")[1]

	d := nip54.NormalizeIdentifier(name)
	evt := nostr.Event{
		Kind:      30818,
		CreatedAt: nostr.Now(),
		Tags: nostr.Tags{
			{"d", d},
			{"title", name},
		},
		Content: def,
	}

	evt.Sign(nostrKey)

	params.logger.Printf("Publishing %s | %s", d, name)

	r, err := params.pool.EnsureRelay(params.relay)
	if err != nil {
		params.logger.Printf("error connecting to relay: %v, retrying in 5 minutes", err)
		time.Sleep(time.Minute * 5)

		return err
	}

	if err := r.Publish(ctx, evt); err != nil {
		params.logger.Printf("error publishing: %v, retrying in 5 minutes", err)
		time.Sleep(time.Minute * 5)

		return err
	}

	return nil
}
