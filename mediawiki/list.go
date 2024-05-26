package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
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

func list(host string) chan string {
	ch := make(chan string)

	apcontinue := os.Getenv("CONTINUE")
	go func() {
		for {
			qs := url.Values{
				"action": {"query"},
				"format": {"json"},
				"list":   {"allpages"},
			}

			if apcontinue != "" {
				qs.Set("apcontinue", apcontinue)
			}

			r, err := http.Get("https://" + host + "/w/api.php?" + qs.Encode())
			if err != nil {
				panic(err)
			}

			var res ListResult
			if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
				r.Body.Close()
				panic(err)
			}
			r.Body.Close()

			for _, page := range res.Query.AllPages {
				ch <- page.Title
			}

			apcontinue = res.Continue.ApContinue
		}
	}()

	return ch
}
