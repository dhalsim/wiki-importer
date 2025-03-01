package mediawiki

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"fiatjaf/wiki-importer/common"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
	"github.com/urfave/cli/v3"
)

func HandleMediaWiki(ctx context.Context, logger *log.Logger, c *cli.Command) error {
	apcontinue := c.String("continue")
	host := c.String("host")

	if host == "" {
		return fmt.Errorf("host is required")
	}

	return runWiki(ctx, logger, apcontinue, host)
}

func runWiki(ctx context.Context, logger *log.Logger, apcontinue, host string) error {
	relayURL, err := common.GetRequiredEnv("RELAY")
	if err != nil {
		return err
	}

	nostrKey, err := common.GetRequiredEnv("NOSTR_KEY")
	if err != nil {
		return err
	}

	pool := nostr.NewSimplePool(ctx)

	if err := CheckWebsiteMatch(
		ctx,
		NewCheckWebsiteMatchParam(
			pool,
			nostrKey,
			host,
			relayURL,
			logger,
		),
	); err != nil {
		return err
	}

	ch, err := getListChannel(host, apcontinue)
	if err != nil {
		return err
	}

	for pageTitle := range ch {
		pageTitle = strings.TrimSpace(pageTitle)

		logger.Println(pageTitle)

		title, asciiDoc, err := asciidoc(host, pageTitle)
		if err != nil {
			logger.Println(err, "\n=========\n-")

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
			Content: asciiDoc,
		}

		if strings.HasPrefix(strings.ToUpper(asciiDoc), ". REDIRECT") {
			// Get the current page's normalized identifier
			currentId := evt.Tags[1][1]

			// Simple extraction between [[ and ]] - more reliable for basic redirects
			parts := strings.Split(asciiDoc, "[[")
			if len(parts) != 2 {
				logger.Println("Could not parse redirect target")

				continue
			}

			targetText := strings.Split(parts[1], "]]")[0]
			target := nip54.NormalizeIdentifier(strings.TrimSpace(targetText))

			// If they normalize to the same identifier, skip
			if target == currentId {
				continue
			}

			evt.Kind = 30819
			evt.Tags = append(evt.Tags, nostr.Tag{"redirect", target})
		}

		evt.Sign(nostrKey)

		relay, err := pool.EnsureRelay(relayURL)
		if err != nil {
			logger.Println(err)

			time.Sleep(time.Second * 2)

			continue
		}

		if err := relay.Publish(ctx, evt); err != nil {
			logger.Println(err)

			time.Sleep(time.Second * 2)

			continue
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}
