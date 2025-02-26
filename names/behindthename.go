package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"fiatjaf/wiki-importer/common"

	"github.com/PuerkitoBio/goquery"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
	"golang.org/x/net/html"
)

func behindthename() {
	pageStart := 1
	if len(os.Args) > 1 {
		pageArg, _ := strconv.Atoi(os.Args[1])
		if pageArg > 0 {
			pageStart = pageArg
		}
	}

	for i := pageStart; i < 101; i++ {
		fmt.Println("PAGE", i)
		proceed := doPage(i)
		if !proceed {
			break
		}
		time.Sleep(time.Second * 5)
	}
}

func doPage(num int) bool {
	var resp *http.Response
	var err error
	for {
		resp, err = common.HttpGet(fmt.Sprintf("https://www.behindthename.com/names/%d", num))
		if err != nil {
			fmt.Println("error page", num, err, "trying again")
			time.Sleep(time.Minute * 5)
			continue
		}
		break
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	sel := doc.Find(`.listname a[href^="/name/"]`)
	if len(sel.Nodes) == 0 {
		return false
	}

	sel.Each(func(_ int, sn *goquery.Selection) {
		href, _ := sn.Attr("href")
		href = strings.TrimSpace(href)
		name := strings.TrimSpace(sn.Text())

		doName("https://www.behindthename.com"+href, name)
	})

	return true
}

func doName(url string, name string) {
	var resp *http.Response
	var err error
	for {
		resp, err = common.HttpGet(url)
		if err != nil {
			fmt.Println("error", url, err, "trying again")
			time.Sleep(5 * time.Minute)
			continue
		}
		break
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	def := ""
	for c := doc.Find(".namedef").Nodes[0].FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			def += c.Data
		case html.ElementNode:
			switch c.Data {
			case "em", "span", "small":
				def += c.FirstChild.Data
			case "a":
				if slices.ContainsFunc(c.Attr, func(a html.Attribute) bool { return a.Key == "href" && strings.HasPrefix(a.Val, "/name/") }) {
					name := strings.TrimSpace(c.FirstChild.Data)
					def += "[[" + name + "]]"
				} else {
					def += c.FirstChild.Data
				}
			case "i":
				def += "_" + c.FirstChild.Data + "_"
			case "b":
				def += "**" + c.FirstChild.Data + "**"
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
	fmt.Println(d, "|", name)

	r, err := pool.EnsureRelay(relay)
	if err != nil {
		fmt.Println("error connecting to relay", err)
		time.Sleep(time.Minute * 5)
		return
	}
	if err := r.Publish(context.Background(), evt); err != nil {
		fmt.Println("error publishing", err)
		time.Sleep(time.Minute * 5)
		return
	}
}
