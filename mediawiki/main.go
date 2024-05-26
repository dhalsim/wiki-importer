package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
	sdk "github.com/nbd-wtf/nostr-sdk"
)

func main() {
	host := os.Getenv("HOST")
	relayURL := os.Getenv("RELAY")
	nostrKey := os.Getenv("NOSTR_KEY")
	if nostrKey == "" || relayURL == "" || host == "" {
		fmt.Println("missing expected environment variables")
		return
	}

	pool := nostr.NewSimplePool(context.Background())
	pub, _ := nostr.GetPublicKey(nostrKey)

	fmt.Printf("[%s] using pubkey=%s relay=%s\n", host, pub, relayURL)

	ctx := context.Background()
	res := pool.QuerySingle(ctx,
		[]string{
			"wss://purplepag.es",
			"wss://relay.nos.social",
			"wss://user.kindpag.es",
		},
		nostr.Filter{Authors: []string{pub}, Kinds: []int{0}, Limit: 1},
	)
	if res == nil {
		fmt.Printf("[%s] no metadata event found for given key\n", host)
		return
	}

	meta, _ := sdk.ParseMetadata(res.Event)
	expected := "https://" + host + "/"
	if meta.Website != expected {
		fmt.Printf("[%s] wrong key: name=%s website=%s, expected website to be %s\n",
			host, meta.Name, meta.Website, expected)
		return
	}

	for s := range list(host) {
		s = strings.TrimSpace(s)
		fmt.Println(s)

		title, md, err := markdown(host, s)
		if err != nil {
			fmt.Println(err, "\n=========\n-")
			time.Sleep(time.Second * 2)
			continue
		}
		title = strings.TrimSpace(title)

		evt := nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      30818,
			Tags: nostr.Tags{
				{"title", title},
				{"d", nip54.NormalizeIdentifier(title)},
			},
			Content: md,
		}

		if strings.HasPrefix(md, "1.  REDIRECT") {
			spl := strings.Split(md, "[[")
			spl = strings.Split(spl[1], "#")
			spl = strings.Split(spl[0], "]]")
			target := nip54.NormalizeIdentifier(strings.TrimSpace(spl[0]))

			if target == evt.Tags[1][1] {
				// this is the same page
				continue
			}

			evt.Kind = 30819
			evt.Tags = append(evt.Tags, nostr.Tag{"redirect", target})
			evt.Content = strings.ReplaceAll(evt.Content, "1.  ", "#")
		}

		evt.Sign(nostrKey)

		relay, err := pool.EnsureRelay(relayURL)
		if err != nil {
			panic(err)
		}

		if err := relay.Publish(ctx, evt); err != nil {
			fmt.Println(err, "\n~~~~~~~~~\n-")
			time.Sleep(time.Second * 2)
			continue
		}

		time.Sleep(2 * time.Second)
	}
}
