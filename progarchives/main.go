package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
)

var logger *log.Logger

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Open log file with append mode, create if doesn't exist
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	logFile, err := os.OpenFile("progarchives-errors.log", flags, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	defer logFile.Close()

	// Create logger that writes to both file and stdout
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)

	relayURL := os.Getenv("RELAY")
	nostrKey := os.Getenv("NOSTR_KEY")
	if nostrKey == "" || relayURL == "" {
		return fmt.Errorf("missing expected environment variables")
	}

	ctx := context.Background()
	pool := nostr.NewSimplePool(context.Background())

	start := 1
	if v, err := strconv.Atoi(os.Getenv("CONTINUE")); err == nil {
		start = v
	}
	var end int

	var fetch func(id int) (string, string, error)
	if len(os.Args) > 1 && os.Args[1] == "albums" {
		end = 75959
		fetch = album
		logger.Printf("Starting album fetch from ID %d to %d\n", start, end)
	} else {
		end = 12736
		fetch = artist
		logger.Printf("Starting artist fetch from ID %d to %d\n", start, end)
	}

	for i := start; i <= end; i++ {
		logger.Printf("Processing ID %d\n", i)

		title, asciiDoc, err := fetch(i)
		if err != nil {
			logger.Printf("Error fetching ID %d: %v\n", i, err)
			time.Sleep(2 * time.Second)
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

		if err := evt.Sign(nostrKey); err != nil {
			logger.Printf("Error signing event for %s: %v\n", title, err)
			continue
		}

		relay, err := pool.EnsureRelay(relayURL)
		if err != nil {
			logger.Printf("Error ensuring relay for %s: %v\n", title, err)
			continue
		}

		if err := relay.Publish(ctx, evt); err != nil {
			logger.Printf("Error publishing %s: %v\n", title, err)
			time.Sleep(time.Second * 2)
			continue
		}

		logger.Printf("Successfully published: %s\n", title)
		time.Sleep(2 * time.Second)
	}

	return nil
}
