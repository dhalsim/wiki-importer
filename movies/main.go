package movies

import (
	"context"
	"embed"
	"fmt"
	"log"
	"text/template"
	"time"

	"fiatjaf/wiki-importer/common"

	"github.com/nbd-wtf/go-nostr"
	"github.com/urfave/cli/v3"
)

const (
	TMDB_MOVIES  = "http://files.tmdb.org/p/exports/movie_ids_%02d_%02d_%d.json.gz"
	TMDB_PERSONS = "http://files.tmdb.org/p/exports/person_ids_%02d_%02d_%d.json.gz"
)

//go:embed tmdb.adoc omdb.adoc person.adoc
var templates embed.FS

var (
	yesterday = time.Now().AddDate(0, 0, -1)
)

func HandleMovies(ctx context.Context, l *log.Logger, c *cli.Command) error {
	startIndex := c.Uint("continue")

	return runMovies(ctx, l, startIndex)
}

func HandlePersons(ctx context.Context, l *log.Logger, c *cli.Command) error {
	startIndex := c.Uint("continue")

	return runPersons(ctx, l, startIndex)
}

func runMovies(ctx context.Context, l *log.Logger, startIndex uint64) error {
	tmdbApiKey, err := common.GetRequiredEnv("TMDB_API_KEY")
	if err != nil {
		return err
	}

	tmdbNostrKey, err := common.GetRequiredEnv("TMDB_NOSTR_KEY")
	if err != nil {
		return err
	}

	tmdbRelay, err := common.GetRequiredEnv("TMDB_RELAY")
	if err != nil {
		return err
	}

	tmdbParsed, err := template.ParseFS(templates, "tmdb.adoc")
	if err != nil {
		return fmt.Errorf("parse TMDB template: %w", err)
	}

	omdbApiKey, err := common.GetRequiredEnv("OMDB_API_KEY")
	if err != nil {
		return err
	}

	omdbNostrKey, err := common.GetRequiredEnv("OMDB_NOSTR_KEY")
	if err != nil {
		return err
	}

	omdbRelay, err := common.GetRequiredEnv("OMDB_RELAY")
	if err != nil {
		return err
	}

	omdbParsed, err := template.ParseFS(templates, "omdb.adoc")
	if err != nil {
		return fmt.Errorf("parse OMDB template: %w", err)
	}

	movies(ctx, MoviesParams{
		Start:        startIndex,
		Pool:         nostr.NewSimplePool(ctx),
		Logger:       l,
		TmdbApiKey:   tmdbApiKey,
		TmdbNostrKey: tmdbNostrKey,
		TmdbRelay:    tmdbRelay,
		TmdbParsed:   tmdbParsed,
		OmdbApiKey:   omdbApiKey,
		OmdbNostrKey: omdbNostrKey,
		OmdbRelay:    omdbRelay,
		OmdbParsed:   omdbParsed,
	})

	return nil
}

func runPersons(ctx context.Context, l *log.Logger, startIndex uint64) error {
	personParsed, err := template.ParseFS(templates, "person.adoc")
	if err != nil {
		return fmt.Errorf("parse person template: %w", err)
	}

	tmdbApiKey, err := common.GetRequiredEnv("TMDB_API_KEY")
	if err != nil {
		return err
	}

	tmdbNostrKey, err := common.GetRequiredEnv("TMDB_NOSTR_KEY")
	if err != nil {
		return err
	}

	tmdbRelay, err := common.GetRequiredEnv("TMDB_RELAY")
	if err != nil {
		return err
	}

	pool := nostr.NewSimplePool(ctx)

	persons(ctx, PersonsParams{
		Start:        startIndex,
		Pool:         pool,
		Logger:       l,
		PersonParsed: personParsed,
		TmdbApiKey:   tmdbApiKey,
		TmdbNostrKey: tmdbNostrKey,
		TmdbRelay:    tmdbRelay,
	})

	return nil
}
