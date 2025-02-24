package main

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

func movies() {
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

	scanner := bufio.NewScanner(gr)
	for scanner.Scan() {
		// TMDB
		imdbId, normalizedIdentifier, err := tmdb(scanner.Bytes())
		if err != nil {
			logger.Printf("Error processing TMDB movie: %v\n", err)
			continue
		} else {
			logger.Printf("Processed TMDB movie: %s, IMDBId: %s\n", normalizedIdentifier, imdbId)
		}

		// OMDB (use the imdb id to query this same movie)
		if err := omdb(imdbId, normalizedIdentifier); err != nil {
			logger.Printf("Error processing OMDB movie: %v\n", err)
			continue
		} else {
			logger.Printf("Processed OMDB movie: %s, IMDBId: %s\n", normalizedIdentifier, imdbId)
		}
	}
}

func tmdb(line []byte) (string, string, error) {
	var movie TMDBMovie
	if err := json.Unmarshal(line, &movie); err != nil {
		return "", "", fmt.Errorf("unmarshal TMDB movie: %w", err)
	}

	{
		logger.Printf("Processing TMDB movie: %d, IMDBId: %s\n", movie.ID, movie.ImdbID)

		// basic movie data
		resp, err := http.Get(fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", movie.ID, tmdbApiKey))
		if err != nil {
			return "", "", fmt.Errorf("fetch TMDB movie details: %w", err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			resp.Body.Close()
			return "", "", fmt.Errorf("decode TMDB movie details: %w", err)
		}
		resp.Body.Close()
		if spl := strings.Split(movie.ReleaseDate, "-"); len(spl) == 3 {
			movie.ReleaseDate = spl[0]
		}
	}

	{
		// cast
		resp, err := http.Get(fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/credits?api_key=%s", movie.ID, tmdbApiKey))
		if err != nil {
			return "", "", fmt.Errorf("fetch TMDB movie credits: %w", err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			resp.Body.Close()
			return "", "", fmt.Errorf("decode TMDB movie credits: %w", err)
		}
		resp.Body.Close()

		if len(movie.Cast) > 4 {
			movie.Cast = movie.Cast[0:4]
		}
	}

	content := &bytes.Buffer{}
	if err := tmdbParsed.Execute(content, movie); err != nil {
		return "", "", fmt.Errorf("execute TMDB template: %w", err)
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
		return "", "", fmt.Errorf("ensure TMDB relay: %w", err)
	}

	if err := relay.Publish(context.Background(), evt); err != nil {
		return "", "", fmt.Errorf("publish to TMDB relay: %w", err)
	}

	return movie.ImdbID, normalizedIdentifier, nil
}

func omdb(imdbId string, normalizedIdentifier string) error {
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

	if err := relay.Publish(context.Background(), evt); err != nil {
		return fmt.Errorf("publish to OMDB relay: %w", err)
	}

	return nil
}
