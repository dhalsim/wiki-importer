package main

import (
	"context"
	"embed"
	"os"
	"text/template"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

const (
	TMDB_MOVIES  = "http://files.tmdb.org/p/exports/movie_ids_01_02_2006.json.gz"
	TMDB_PERSONS = "http://files.tmdb.org/p/exports/person_ids_01_02_2006.json.gz"
)

//go:embed tmdb.adoc omdb.adoc
var templates embed.FS

var (
	tmdbApiKey   string
	tmdbNostrKey string
	tmdbRelay    string

	tmdbParsed *template.Template

	omdbApiKey   string
	omdbNostrKey string
	omdbRelay    string

	omdbParsed *template.Template

	now  = time.Now().Add(time.Hour * -3)
	pool *nostr.SimplePool
)

func main() {
	tmdbApiKey = os.Getenv("TMDB_API_KEY")
	tmdbNostrKey = os.Getenv("TMDB_NOSTR_KEY")
	tmdbRelay = os.Getenv("TMDB_RELAY")

	var err error
	tmdbParsed, err = template.ParseFS(templates, "tmdb.adoc")
	if err != nil {
		panic(err)
	}

	omdbApiKey = os.Getenv("OMDB_API_KEY")
	omdbNostrKey = os.Getenv("OMDB_NOSTR_KEY")
	omdbRelay = os.Getenv("OMDB_RELAY")

	omdbParsed, err = template.ParseFS(templates, "omdb.adoc")
	if err != nil {
		panic(err)
	}

	pool = nostr.NewSimplePool(context.Background())

	movies()
	persons()
}
