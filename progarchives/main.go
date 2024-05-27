package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
)

func main() {
	relayURL := os.Getenv("RELAY")
	nostrKey := os.Getenv("NOSTR_KEY")
	if nostrKey == "" || relayURL == "" {
		fmt.Println("missing expected environment variables")
		return
	}

	ctx := context.Background()
	pool := nostr.NewSimplePool(context.Background())

	start := 0 // for some reason the first request will always fail, so do it on id 0, which doesn't exist
	if v, err := strconv.Atoi(os.Getenv("CONTINUE")); err == nil {
		start = v
	}
	var end int

	var fetch func(id int) (string, string, error)
	if len(os.Args) > 1 && os.Args[1] == "albums" {
		end = 75959
		fetch = album
	} else {
		end = 12736
		fetch = artist
	}

	for i := start; i <= end; i++ {
		fmt.Println("id", i)
		title, md, err := fetch(i)
		if err != nil {
			fmt.Println(err, "\n~~~~~~~~~\n-")
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("  ", title)
		evt := nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      30818,
			Tags: nostr.Tags{
				{"title", title},
				{"d", nip54.NormalizeIdentifier(title)},
			},
			Content: md,
		}

		if err := evt.Sign(nostrKey); err != nil {
			panic(err)
		}

		relay, err := pool.EnsureRelay(relayURL)
		if err != nil {
			panic(err)
		}

		if err := relay.Publish(ctx, evt); err != nil {
			fmt.Println(err, "\n~~~~~~~~~\n-")
			fmt.Println(evt)
			time.Sleep(time.Second * 2)
			continue
		}

		time.Sleep(2 * time.Second)
	}
}
