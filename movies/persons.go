package movies

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"text/template"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip54"

	"fiatjaf/wiki-importer/common"
)

type PersonsParams struct {
	Start        uint64
	Pool         *nostr.SimplePool
	Logger       *log.Logger
	PersonParsed *template.Template
	TmdbApiKey   string
	TmdbNostrKey string
	TmdbRelay    string
}

func persons(ctx context.Context, params PersonsParams) {
	start := params.Start
	pool := params.Pool
	logger := params.Logger
	personParsed := params.PersonParsed
	tmdbApiKey := params.TmdbApiKey
	tmdbNostrKey := params.TmdbNostrKey
	tmdbRelay := params.TmdbRelay

	resp, err := common.HttpGet(getYesterdays(TMDB_PERSONS))
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

		i++

		var person TMDBPerson
		if err := json.Unmarshal(scanner.Bytes(), &person); err != nil {
			logger.Printf("Error unmarshalling TMDB person - index: %d, %v\n", i, err)

			continue
		}

		searchURL := fmt.Sprintf(
			"https://api.themoviedb.org/3/search/person?query=%s&api_key=%s",
			url.QueryEscape(person.Name),
			tmdbApiKey,
		)

		logger.Printf("Fetching TMDB person - index: %d, name: %s\n", i, person.Name)

		resp, err := common.HttpGet(searchURL)
		if err != nil {
			logger.Printf("Error fetching TMDB person - index: %d, %v\n", i, err)

			continue
		}

		if resp.StatusCode != 200 {
			logger.Printf(
				"Error fetching TMDB person - index: %d, status code: %d\n",
				i,
				resp.StatusCode,
			)

			resp.Body.Close()

			// Add a small delay to respect rate limits
			time.Sleep(1 * time.Second)

			continue
		}

		var apiResult TMDBPersonApiResult
		if err := json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
			body, _ := io.ReadAll(resp.Body)
			logger.Printf("Error decoding TMDB person - index: %d, %v\n", i, err)
			logger.Printf("Response body: %s\n", string(body))
			resp.Body.Close()

			continue
		}
		resp.Body.Close()

		if len(apiResult.Results) == 0 {
			logger.Printf("No results found for TMDB person - index: %d, name: %s\n", i, person.Name)

			continue
		}

		// Process each result
		for resultIndex, result := range apiResult.Results {
			logger.Printf(
				"Found match for TMDB person - ID: %d, Name: %s, index %d, result index: %d\n",
				result.ID,
				result.Name,
				i,
				resultIndex,
			)

			content := &bytes.Buffer{}
			if err := personParsed.Execute(content, result); err != nil {
				logger.Printf(
					"Error executing TMDB template - index: %d, result index: %d, %v\n",
					i,
					resultIndex,
					err,
				)

				continue
			}

			normalizedIdentifier := nip54.NormalizeIdentifier(result.Name)

			evt := nostr.Event{
				CreatedAt: nostr.Now(),
				Kind:      30818,
				Tags: nostr.Tags{
					{"title", result.Name},
					{"d", normalizedIdentifier},
				},
				Content: content.String(),
			}

			evt.Sign(tmdbNostrKey)

			relay, err := pool.EnsureRelay(tmdbRelay)
			if err != nil {
				logger.Printf(
					"Error ensuring TMDB relay - index: %d, result index: %d, %v\n",
					i,
					resultIndex,
					err,
				)

				continue
			}

			if err := relay.Publish(ctx, evt); err != nil {
				logger.Printf(
					"Error publishing TMDB person - index: %d, result index: %d, %v\n",
					i,
					resultIndex,
					err,
				)
			}
		}

		// Add a small delay between requests to respect rate limits
		time.Sleep(1 * time.Second)
	}
}
