package movies

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
)

type TMDBResult struct {
	TMDBId               int
	IMDBId               string
	NormalizedIdentifier string
}

type TmdbParams struct {
	Index        uint64
	Line         []byte
	Logger       *log.Logger
	TmdbApiKey   string
	TmdbNostrKey string
	TmdbRelay    string
	Pool         *nostr.SimplePool
	TmdbParsed   *template.Template
}

func NewTmdbParams(
	index uint64,
	line []byte,
	logger *log.Logger,
	tmdbApiKey string,
	tmdbNostrKey string,
	tmdbRelay string,
	pool *nostr.SimplePool,
	tmdbParsed *template.Template,
) TmdbParams {
	return TmdbParams{
		Index:        index,
		Line:         line,
		Logger:       logger,
		TmdbApiKey:   tmdbApiKey,
		TmdbNostrKey: tmdbNostrKey,
		TmdbRelay:    tmdbRelay,
		Pool:         pool,
		TmdbParsed:   tmdbParsed,
	}
}

func tmdb(ctx context.Context, params TmdbParams) (TMDBResult, error) {
	index := params.Index
	line := params.Line
	logger := params.Logger
	tmdbApiKey := params.TmdbApiKey
	tmdbNostrKey := params.TmdbNostrKey
	tmdbRelay := params.TmdbRelay
	pool := params.Pool
	tmdbParsed := params.TmdbParsed

	empty := TMDBResult{}

	var movie TMDBMovie
	if err := json.Unmarshal(line, &movie); err != nil {
		return empty, fmt.Errorf("unmarshal TMDB movie - index: %d, %w", index, err)
	}

	{
		logger.Printf("Processing TMDB movie - ID: %d, index: %d\n", movie.ID, index)

		// basic movie data
		resp, err := http.Get(
			fmt.Sprintf(
				"https://api.themoviedb.org/3/movie/%d?api_key=%s",
				movie.ID,
				tmdbApiKey,
			),
		)

		if err != nil {
			return empty, fmt.Errorf("fetch TMDB movie details: %w", err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			resp.Body.Close()
			return empty, fmt.Errorf("decode TMDB movie details: %w", err)
		}
		resp.Body.Close()

		if spl := strings.Split(movie.ReleaseDate, "-"); len(spl) == 3 {
			movie.ReleaseDate = spl[0]
		}
	}

	{
		// cast
		resp, err := http.Get(
			fmt.Sprintf(
				"https://api.themoviedb.org/3/movie/%d/credits?api_key=%s",
				movie.ID,
				tmdbApiKey,
			),
		)

		if err != nil {
			return empty, fmt.Errorf("fetch TMDB movie credits: %w", err)
		}

		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			resp.Body.Close()
			return empty, fmt.Errorf("decode TMDB movie credits: %w", err)
		}
		resp.Body.Close()

		if len(movie.Cast) > 4 {
			movie.Cast = movie.Cast[0:4]
		}
	}

	content := &bytes.Buffer{}
	if err := tmdbParsed.Execute(content, movie); err != nil {
		return empty, fmt.Errorf("execute TMDB template: %w", err)
	}

	normalizedIdentifier := nip54.NormalizeIdentifier(movie.Title)

	evt := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      30818,
		Tags: nostr.Tags{
			{"title", movie.Title},
			{"d", normalizedIdentifier},
		},
		Content: content.String(),
	}

	evt.Sign(tmdbNostrKey)

	relay, err := pool.EnsureRelay(tmdbRelay)
	if err != nil {
		return empty, fmt.Errorf("error ensuring TMDB relay - index: %d, %w", index, err)
	}

	if err := relay.Publish(ctx, evt); err != nil {
		return empty, fmt.Errorf("error publishing TMDB event - index: %d, %w", index, err)
	}

	return TMDBResult{
		TMDBId:               movie.ID,
		IMDBId:               movie.ImdbID,
		NormalizedIdentifier: normalizedIdentifier,
	}, nil
}
