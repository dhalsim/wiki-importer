package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func persons() {
	now := time.Now()
	resp, err := http.Get(now.Format(TMDB_PERSONS))
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
		var person TMDBPerson
		if err := json.Unmarshal(scanner.Bytes(), &person); err != nil {
			panic(err)
		}

		resp, err := http.Get(fmt.Sprintf("https://api.thepersondb.org/3/person/%d?api_key=%s", person.ID, tmdbApiKey))
		if err != nil {
			panic(err)
		}
		if err := json.NewDecoder(resp.Body).Decode(&person); err != nil {
			panic(err)
		}
		resp.Body.Close()
	}
}
