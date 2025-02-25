package movies

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"
)

func movies(ctx context.Context, start uint64) {
	pool = nostr.NewSimplePool(ctx)

	resp, err := http.Get(now.Format(TMDB_MOVIES))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(err)
	}
	defer gr.Close()

	i := uint64(0)
	scanner := bufio.NewScanner(gr)
	for scanner.Scan() {
		if i < start {
			i++
			continue
		}

		// TMDB
		tmdbResult, err := tmdb(ctx, i, scanner.Bytes())
		if err != nil {
			logger.Printf("Error processing TMDB movie: %v\n", err)
			continue
		} else {
			logger.Printf(
				"Processed TMDB movie - ID: %d, %s, IMDBId: %s, index: %d\n",
				tmdbResult.TMDBId,
				tmdbResult.NormalizedIdentifier,
				tmdbResult.IMDBId,
				i,
			)
		}

		// OMDB (use the imdb id to query this same movie also the same normalized identifier)
		if err := omdb(ctx, tmdbResult.IMDBId, tmdbResult.NormalizedIdentifier); err != nil {
			logger.Printf("Error processing OMDB movie: %v\n", err)
			continue
		} else {
			logger.Printf(
				"Processed OMDB movie: %s, IMDBId: %s, index: %d\n",
				tmdbResult.NormalizedIdentifier,
				tmdbResult.IMDBId,
				i,
			)
		}

		i++
	}
}

type TMDBResult struct {
	TMDBId               int
	IMDBId               string
	NormalizedIdentifier string
}

func tmdb(ctx context.Context, index uint64, line []byte) (TMDBResult, error) {
	var empty = TMDBResult{}

	var movie TMDBMovie
	if err := json.Unmarshal(line, &movie); err != nil {
		return empty, fmt.Errorf("unmarshal TMDB movie: %w", err)
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
		return empty, fmt.Errorf("ensure TMDB relay: %w", err)
	}

	if err := relay.Publish(ctx, evt); err != nil {
		return empty, fmt.Errorf("publish to TMDB relay: %w", err)
	}

	return TMDBResult{
		TMDBId:               movie.ID,
		IMDBId:               movie.ImdbID,
		NormalizedIdentifier: normalizedIdentifier,
	}, nil
}

func omdb(ctx context.Context, imdbId string, normalizedIdentifier string) error {
	var movie OMDBMovie

	resp, err := http.Get(fmt.Sprintf("https://www.omdbapi.com/?i=%s&plot=full&apikey=%s", imdbId, omdbApiKey))
	if err != nil {
		return fmt.Errorf("fetch OMDB movie: %w", err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		resp.Body.Close()
		return fmt.Errorf("decode OMDB movie: %w", err)
	}
	resp.Body.Close()

	logger.Printf("Processing OMDB movie: %s, IMDBId: %s\n", normalizedIdentifier, imdbId)

	movie.Director = splitAndWikilink(movie.Director)
	movie.Writer = splitAndWikilink(movie.Writer)
	movie.Actors = splitAndWikilink(movie.Actors)
	movie.Genre = splitAndWikilink(movie.Genre)

	content := &bytes.Buffer{}
	if err := omdbParsed.Execute(content, movie); err != nil {
		return fmt.Errorf("execute OMDB template: %w", err)
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
		return fmt.Errorf("ensure OMDB relay: %w", err)
	}

	if err := relay.Publish(ctx, evt); err != nil {
		return fmt.Errorf("publish to OMDB relay: %w", err)
	}

	return nil
}
