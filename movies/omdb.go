package movies

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/nbd-wtf/go-nostr"
)

type OmdbParams struct {
	Index                uint64
	ImdbId               string
	NormalizedIdentifier string
	OmdbApiKey           string
	OmdbNostrKey         string
	OmdbRelay            string
	Pool                 *nostr.SimplePool
	Logger               *log.Logger
	OmdbParsed           *template.Template
}

func NewOmdbParams(
	index uint64,
	imdbId string,
	normalizedIdentifier string,
	omdbApiKey string,
	omdbNostrKey string,
	omdbRelay string,
	pool *nostr.SimplePool,
	logger *log.Logger,
	omdbParsed *template.Template,
) OmdbParams {
	return OmdbParams{
		Index:                index,
		ImdbId:               imdbId,
		NormalizedIdentifier: normalizedIdentifier,
		OmdbApiKey:           omdbApiKey,
		OmdbNostrKey:         omdbNostrKey,
		OmdbRelay:            omdbRelay,
		Pool:                 pool,
		Logger:               logger,
		OmdbParsed:           omdbParsed,
	}
}

func omdb(ctx context.Context, params OmdbParams) error {
	index := params.Index
	imdbId := params.ImdbId
	normalizedIdentifier := params.NormalizedIdentifier
	omdbApiKey := params.OmdbApiKey
	omdbNostrKey := params.OmdbNostrKey
	omdbRelay := params.OmdbRelay
	pool := params.Pool
	logger := params.Logger
	omdbParsed := params.OmdbParsed

	resp, err := http.Get(fmt.Sprintf("https://www.omdbapi.com/?i=%s&plot=full&apikey=%s", imdbId, omdbApiKey))
	if err != nil {
		return fmt.Errorf("error fetching OMDB movie - index: %d, %w", index, err)
	}

	var movie OMDBMovie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		resp.Body.Close()
		return fmt.Errorf("error decoding OMDB movie - index: %d, %w", index, err)
	}
	resp.Body.Close()

	logger.Printf(
		"Processing OMDB movie - index: %d, %s, IMDBId: %s\n",
		index,
		normalizedIdentifier,
		imdbId,
	)

	movie.Director = splitAndWikilink(movie.Director)
	movie.Writer = splitAndWikilink(movie.Writer)
	movie.Actors = splitAndWikilink(movie.Actors)
	movie.Genre = splitAndWikilink(movie.Genre)

	content := &bytes.Buffer{}
	if err := omdbParsed.Execute(content, movie); err != nil {
		return fmt.Errorf("error executing OMDB template - index: %d, %w", index, err)
	}

	evt := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      30818,
		Tags: nostr.Tags{
			{"title", movie.Title},
			{"d", normalizedIdentifier},
		},
		Content: content.String(),
	}

	evt.Sign(omdbNostrKey)

	relay, err := pool.EnsureRelay(omdbRelay)
	if err != nil {
		return fmt.Errorf("error ensuring OMDB relay - index: %d, %w", index, err)
	}

	if err := relay.Publish(ctx, evt); err != nil {
		return fmt.Errorf("error publishing OMDB event - index: %d, %w", index, err)
	}

	return nil
}
