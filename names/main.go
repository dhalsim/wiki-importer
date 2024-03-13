package main

import (
	"context"
	"os"

	"github.com/nbd-wtf/go-nostr"
)

var (
	pool     *nostr.SimplePool
	nostrKey string
	relay    string
)

func main() {
	nostrKey = os.Getenv("BEHINDTHENAME_NOSTR_KEY")
	relay = os.Getenv("BEHINDTHENAME_RELAY")

	pool = nostr.NewSimplePool(context.Background())
	behindthename()
}
