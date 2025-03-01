package mediawiki

import (
	"context"
	"fmt"
	"log"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/sdk"
)

type CheckWebsiteMatchParam struct {
	Pool     *nostr.SimplePool
	nostrKey string
	host     string
	relayURL string
	logger   *log.Logger
}

func NewCheckWebsiteMatchParam(
	pool *nostr.SimplePool,
	nostrKey,
	host,
	relayURL string,
	logger *log.Logger,
) CheckWebsiteMatchParam {
	return CheckWebsiteMatchParam{
		Pool:     pool,
		nostrKey: nostrKey,
		host:     host,
		relayURL: relayURL,
		logger:   logger,
	}
}

func CheckWebsiteMatch(ctx context.Context, params CheckWebsiteMatchParam) error {
	logger := params.logger
	host := params.host
	relayURL := params.relayURL

	pub, _ := nostr.GetPublicKey(params.nostrKey)

	logger.Printf("[%s] using pubkey=%s relay=%s\n", host, pub, relayURL)

	res := params.Pool.QuerySingle(ctx,
		[]string{
			"wss://purplepag.es",
			"wss://relay.nos.social",
			"wss://user.kindpag.es",
		},
		nostr.Filter{Authors: []string{pub}, Kinds: []int{0}, Limit: 1},
	)
	if res == nil {
		return fmt.Errorf("[%s] no metadata event found for given key", host)
	}

	meta, _ := sdk.ParseMetadata(res.Event)
	expected := "https://" + params.host + "/"

	if meta.Website != expected {
		return fmt.Errorf(
			"[%s] wrong key: name=%s website=%s, expected website to be %s",
			params.host,
			meta.Name,
			meta.Website,
			expected,
		)
	}

	return nil
}
