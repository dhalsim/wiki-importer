package movies

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/urfave/cli/v3"
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

func HandleMovies(ctx context.Context, l *log.Logger, c *cli.Command) error {
	startIndex := c.Uint("continue")
	logger = l

	return runMovies(ctx, startIndex)
}

func HandlePersons(ctx context.Context, l *log.Logger, c *cli.Command) error {
	startIndex := c.Uint("continue")
	logger = l

	return runPersons(ctx, startIndex)
}

func runMovies(ctx context.Context, startIndex uint64) error {
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

	var err error
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

	movies(ctx, startIndex)

	return nil
}

func runPersons(ctx context.Context, startIndex uint64) error {
	persons(ctx, startIndex)

	return nil
}
