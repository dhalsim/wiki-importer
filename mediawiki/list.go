package mediawiki

import (
	"encoding/json"
	"net/url"

	"fiatjaf/wiki-importer/common"
)

type ListResult struct {
	Continue struct {
		ApContinue string `json:"apcontinue"`
	} `json:"continue"`
	Query struct {
		AllPages []struct {
			PageID int    `json:"pageid"`
			Title  string `json:"title"`
		} `json:"allpages"`
	} `json:"query"`
}

func getListChannel(host string, apcontinue string) (chan string, error) {
	ch := make(chan string)
	errCh := make(chan error, 1) // buffered channel for errors

	go func() {
		defer close(ch)
		defer close(errCh)

		for {
			qs := url.Values{
				"action": {"query"},
				"format": {"json"},
				"list":   {"allpages"},
			}

			if apcontinue != "" {
				qs.Set("apcontinue", apcontinue)
			}

			r, err := common.HttpGet(apiBase(host) + "?" + qs.Encode())
			if err != nil {
				errCh <- err

				return
			}

			var res ListResult
			if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
				r.Body.Close()
				errCh <- err

				return
			}
			r.Body.Close()

			for _, page := range res.Query.AllPages {
				ch <- page.Title
			}

			if res.Continue.ApContinue == "" {
				// No more pages to fetch
				return
			}

			apcontinue = res.Continue.ApContinue
		}
	}()

	// Check for immediate errors
	select {
	case err := <-errCh:
		close(ch)

		return nil, err
	default:
		return ch, nil
	}
}
