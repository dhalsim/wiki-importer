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
		imdbId := tmdb(scanner.Bytes())

		// OMDB (use the imdb id to query this same movie)
		omdb(imdbId)
	}
}

func tmdb(line []byte) string {
	var movie TMDBMovie
	if err := json.Unmarshal(line, &movie); err != nil {
		panic(err)
	}

	{
		// basic movie data
		resp, err := http.Get(fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", movie.ID, tmdbApiKey))
		if err != nil {
			panic(err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			panic(err)
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
			panic(err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			panic(err)
		}
		resp.Body.Close()

		if len(movie.Cast) > 4 {
			movie.Cast = movie.Cast[0:4]
		}
	}

	content := &bytes.Buffer{}
	if err := tmdbParsed.Execute(content, movie); err != nil {
		panic(err)
	}

	evt := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      30818,
		Tags: nostr.Tags{
			{"title", movie.Title},
			{"d", nip54.NormalizeIdentifier(movie.Title)},
		},
		Content: content.String(),
	}
	evt.Sign(tmdbNostrKey)

	relay, err := pool.EnsureRelay(tmdbRelay)
	if err != nil {
		panic(err)
	}

	if err := relay.Publish(context.Background(), evt); err != nil {
		panic(err)
	}

	return movie.ImdbID
}

func omdb(imdbId string) {
	var movie OMDBMovie

	resp, err := http.Get(fmt.Sprintf("https://www.omdbapi.com/?i=%s&plot=full&apikey=%s", imdbId, tmdbApiKey))
	if err != nil {
		panic(err)
	}
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		panic(err)
	}
	resp.Body.Close()

	movie.Director = splitAndWikilink(movie.Director)
	movie.Writer = splitAndWikilink(movie.Director)
	movie.Actors = splitAndWikilink(movie.Director)
	movie.Genre = splitAndWikilink(movie.Director)

	content := &bytes.Buffer{}
	if err := omdbParsed.Execute(content, movie); err != nil {
		panic(err)
	}

	evt := nostr.Event{
		CreatedAt: nostr.Now(),
		Kind:      30818,
		Tags: nostr.Tags{
			{"title", movie.Title},
			{"d", nip54.NormalizeIdentifier(movie.Title)},
		},
		Content: content.String(),
	}
	evt.Sign(omdbNostrKey)

	relay, err := pool.EnsureRelay(omdbRelay)
	if err != nil {
		panic(err)
	}

	if err := relay.Publish(context.Background(), evt); err != nil {
		panic(err)
	}
}
