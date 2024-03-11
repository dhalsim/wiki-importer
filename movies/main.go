package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

const (
	TMDB_MOVIES  = "http://files.tmdb.org/p/exports/movie_ids_01_02_2006.json.gz"
	TMDB_PERSONS = "http://files.tmdb.org/p/exports/person_ids_01_02_2006.json.gz"
)

func main() {
	now := time.Now()
	tmdbApiKey := os.Getenv("TMDB_API_KEY")

	fmt.Println(now.Format(TMDB_MOVIES))
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
		var movie TMDBMovie
		if err := json.Unmarshal(scanner.Bytes(), &movie); err != nil {
			panic(err)
		}

		fmt.Println(movie.ID)
		resp, err := http.Get(fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s", movie.ID, tmdbApiKey))
		if err != nil {
			panic(err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
			panic(err)
		}
		resp.Body.Close()

		evt := nostr.Event{
			CreatedAt: nostr.Now(),
			Kind:      30818,
			Tags:      nostr.Tags{},
		}
	}
}
