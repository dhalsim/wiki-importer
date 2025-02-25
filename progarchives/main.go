package progarchives

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
	"github.com/urfave/cli/v3"
)

var logger *log.Logger

type CommonConfig struct {
	Pool     *nostr.SimplePool // Nostr connection pool
	Ctx      context.Context   // Application context
	NostrKey string            // Nostr private key
	RelayURL string            // Nostr relay URL
}

func common(ctx context.Context) (*CommonConfig, error) {
	relayURL := os.Getenv("RELAY")
	if relayURL == "" {
		return nil, fmt.Errorf("RELAY environment variable is required")
	}

	nostrKey := os.Getenv("NOSTR_KEY")
	if nostrKey == "" {
		return nil, fmt.Errorf("NOSTR_KEY environment variable is required")
	}

	pool := nostr.NewSimplePool(ctx)

	return &CommonConfig{
		Pool:     pool,
		Ctx:      ctx,
		NostrKey: nostrKey,
		RelayURL: relayURL,
	}, nil
}

// FetchFunc represents a function that fetches data by ID
type FetchFunc func(id uint64) (title string, asciiDoc string, err error)

type RunParams struct {
	Start uint64
	End   uint64
	Fetch FetchFunc
}

func run(ctx context.Context, params *RunParams) error {
	cfg, err := common(ctx)
	if err != nil {
		return err
	}

	for i := params.Start; i <= params.End; i++ {
		logger.Printf("Processing ID %d\n", i)

		title, asciiDoc, err := params.Fetch(i)
		if err != nil {
			logger.Printf("Error fetching ID %d: %v\n", i, err)

			time.Sleep(2 * time.Second)

			// Try with the next one
			continue
		}

		logger.Printf("Successfully fetched: %s\n", title)

		evt := nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      30818,
			Tags: nostr.Tags{
				{"title", title},
				{"d", nip54.NormalizeIdentifier(title)},
			},
			Content: asciiDoc,
		}

		if err := evt.Sign(cfg.NostrKey); err != nil {
			logger.Printf("Error signing event for %s: %v\n", title, err)

			// Try with the next one
			continue
		}

		relay, err := cfg.Pool.EnsureRelay(cfg.RelayURL)
		if err != nil {
			logger.Printf("Error ensuring relay for %s: %v\n", title, err)

			// Try with the next one
			continue
		}

		if err := relay.Publish(cfg.Ctx, evt); err != nil {
			logger.Printf("Error publishing %s: %v\n", title, err)

			// Try with the next one
			continue
		}

		logger.Printf("Successfully published: %s\n", title)

		time.Sleep(2 * time.Second)
	}

	return nil
}

func HandleAlbums(ctx context.Context, l *log.Logger, c *cli.Command) error {
	logger = l

	return run(ctx, &RunParams{
		Start: c.Uint("continue"),
		End:   75959,
		Fetch: album,
	})
}

func HandleArtists(ctx context.Context, l *log.Logger, c *cli.Command) error {
	logger = l

	return run(ctx, &RunParams{
		Start: c.Uint("continue"),
		End:   12736,
		Fetch: artist,
	})
}
