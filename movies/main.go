package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
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

	logger *log.Logger
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Open log file with append mode, create if doesn't exist
	logFile, err := os.OpenFile("movies-errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	defer logFile.Close()

	// Create logger that writes to both file and stdout
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags)

	tmdbApiKey = os.Getenv("TMDB_API_KEY")
	if tmdbApiKey == "" {
		return fmt.Errorf("TMDB_API_KEY environment variable is required")
	}

	tmdbNostrKey = os.Getenv("TMDB_NOSTR_KEY")
	if tmdbNostrKey == "" {
		return fmt.Errorf("TMDB_NOSTR_KEY environment variable is required")
	}

	tmdbRelay = os.Getenv("TMDB_RELAY")
	if tmdbRelay == "" {
		return fmt.Errorf("TMDB_RELAY environment variable is required")
	}

	tmdbParsed, err = template.ParseFS(templates, "tmdb.adoc")
	if err != nil {
		return fmt.Errorf("parse TMDB template: %w", err)
	}

	omdbApiKey = os.Getenv("OMDB_API_KEY")
	if omdbApiKey == "" {
		return fmt.Errorf("OMDB_API_KEY environment variable is required")
	}

	omdbNostrKey = os.Getenv("OMDB_NOSTR_KEY")
	if omdbNostrKey == "" {
		return fmt.Errorf("OMDB_NOSTR_KEY environment variable is required")
	}

	omdbRelay = os.Getenv("OMDB_RELAY")
	if omdbRelay == "" {
		return fmt.Errorf("OMDB_RELAY environment variable is required")
	}

	omdbParsed, err = template.ParseFS(templates, "omdb.adoc")
	if err != nil {
		return fmt.Errorf("parse OMDB template: %w", err)
	}

	pool = nostr.NewSimplePool(context.Background())

	movies()
	persons()

	return nil
}
