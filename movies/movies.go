package movies

import (
	"bufio"
	"compress/gzip"
	"context"
	"log"
	"text/template"

	"fiatjaf/wiki-importer/common"

	"github.com/nbd-wtf/go-nostr"
)

type MoviesParams struct {
	Start        uint64
	Pool         *nostr.SimplePool
	Logger       *log.Logger
	TmdbApiKey   string
	TmdbNostrKey string
	TmdbRelay    string
	TmdbParsed   *template.Template
	OmdbApiKey   string
	OmdbNostrKey string
	OmdbRelay    string
	OmdbParsed   *template.Template
}

func movies(ctx context.Context, params MoviesParams) {
	start := params.Start
	tmdbApiKey := params.TmdbApiKey
	tmdbNostrKey := params.TmdbNostrKey
	tmdbRelay := params.TmdbRelay
	omdbApiKey := params.OmdbApiKey
	omdbNostrKey := params.OmdbNostrKey
	omdbRelay := params.OmdbRelay
	omdbParsed := params.OmdbParsed
	pool := params.Pool
	logger := params.Logger
	tmdbParsed := params.TmdbParsed

	resp, err := common.HttpGet(getYesterdays(TMDB_MOVIES))
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
		tmdbResult, err := tmdb(ctx, NewTmdbParams(
			i,
			scanner.Bytes(),
			logger,
			tmdbApiKey,
			tmdbNostrKey,
			tmdbRelay,
			pool,
			tmdbParsed,
		))
		if err != nil {
			logger.Printf("Error processing TMDB movie - index: %d, %v\n", i, err)

			continue
		}
		logger.Printf(
			"Processed TMDB movie - ID: %d, %s, IMDBId: %s, index: %d\n",
			tmdbResult.TMDBId,
			tmdbResult.NormalizedIdentifier,
			tmdbResult.IMDBId,
			i,
		)

		// OMDB
		if err := omdb(ctx, NewOmdbParams(
			i,
			tmdbResult.IMDBId,
			tmdbResult.NormalizedIdentifier,
			omdbApiKey,
			omdbNostrKey,
			omdbRelay,
			pool,
			logger,
			omdbParsed,
		)); err != nil {
			logger.Printf("Error processing OMDB movie - index: %d, %v\n", i, err)

			continue
		}

		i++
	}
}
