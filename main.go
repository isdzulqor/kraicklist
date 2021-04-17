package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"kraicklist/domain/model"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// initialize searcher
	searcher := &Searcher{}
	err := searcher.Load("./data/data.gz")
	if err != nil {
		log.Fatalf("unable to load search data due: %v", err)
	}
	// define http handlers
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/search", handleSearch(searcher))
	// define port, we need to set it as env for Heroku deployment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	// start server
	fmt.Printf("Server is listening on %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatalf("unable to start server due: %v", err)
	}
}

func handleSearch(s *Searcher) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// fetch query string from query params
			q := r.URL.Query().Get("q")
			if len(q) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("missing search query in query params"))
				return
			}
			// search relevant ads
			ads, err := s.Search(q)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			// output success response
			buf := new(bytes.Buffer)
			encoder := json.NewEncoder(buf)
			encoder.Encode(ads)
			w.Header().Set("Content-Type", "application/json")
			w.Write(buf.Bytes())
		},
	)
}

type Searcher struct {
	ads model.Advertisements
}

func (s *Searcher) Load(filepath string) error {
	// open file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("unable to open source file due: %v", err)
	}
	defer file.Close()
	// read as gzip
	reader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("unable to initialize gzip reader due: %v", err)
	}
	// read the reader using scanner to contstruct ads
	var ads model.Advertisements
	cs := bufio.NewScanner(reader)
	for cs.Scan() {
		var r model.Advertisement
		err = json.Unmarshal(cs.Bytes(), &r)
		if err != nil {
			continue
		}
		ads = append(ads, r)
	}
	s.ads = ads

	return nil
}

func (s *Searcher) Search(query string) (model.Advertisements, error) {
	var result model.Advertisements
	for _, record := range s.ads {
		if strings.Contains(record.Title, query) || strings.Contains(record.Content, query) {
			result = append(result, record)
		}
	}
	return result, nil
}
