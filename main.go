package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"fiatjaf/wiki-importer/movies"
	"fiatjaf/wiki-importer/progarchives"

	"github.com/urfave/cli/v3"
)

func main() {
	continueFlag := &cli.UintFlag{
		Name:    "continue",
		Aliases: []string{"c"},
		Usage:   "Continue from specific index",
		Value:   0, // default value
	}

	cmd := &cli.Command{
		Name:  "wiki-importer",
		Usage: "Import data from various sources and publish to Nostr as NIP-54 Wiki content",
		Commands: []*cli.Command{
			{
				Name:  "progarchives",
				Usage: "Import data from progarchives",
				Commands: []*cli.Command{
					{
						Name:  "albums",
						Usage: "Import albums from progarchives",
						Flags: []cli.Flag{
							continueFlag,
						},
						Action: handleProgArchivesAlbums,
					},
					{
						Name:  "artists",
						Usage: "Import artists from progarchives",
						Flags: []cli.Flag{
							continueFlag,
						},
						Action: handleProgArchivesArtists,
					},
				},
			},
			{
				Name:  "movies",
				Usage: "Import data from TMDB and OMDB",
				Flags: []cli.Flag{
					continueFlag,
				},
				Action: handleMovies,
				Commands: []*cli.Command{
					{
						Name:  "persons",
						Usage: "Import persons from The Person DB",
						Flags: []cli.Flag{
							continueFlag,
						},
						Action: handlePersons,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func createLogger(name string) (*log.Logger, error) {
	const logFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("create logs directory: %w", err)
	}

	logFile, err := os.OpenFile("logs/"+name+".log", logFlags, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	return log.New(io.MultiWriter(os.Stdout, logFile), "", log.LstdFlags), nil
}

func handleProgArchivesAlbums(ctx context.Context, c *cli.Command) error {
	logger, err := createLogger("progarchives-albums")
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	return progarchives.HandleAlbums(ctx, logger, c)
}

func handleProgArchivesArtists(ctx context.Context, c *cli.Command) error {
	logger, err := createLogger("progarchives-artists")
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	return progarchives.HandleArtists(ctx, logger, c)
}

func handleMovies(ctx context.Context, c *cli.Command) error {
	logger, err := createLogger("movies")
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	return movies.HandleMovies(ctx, logger, c)
}

func handlePersons(ctx context.Context, c *cli.Command) error {
	logger, err := createLogger("movies-persons")
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	return movies.HandlePersons(ctx, logger, c)
}
