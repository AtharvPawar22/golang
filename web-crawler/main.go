package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

// Fetcher interface to allow mocking later
type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

// HttpFetcher implements Fetcher using net/http
type HttpFetcher struct{}

func (f HttpFetcher) Fetch(url string) (string, []string, error) {
	fmt.Printf("Fetching: %s\n", url)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read body into a reader that can be reused if needed, 
	// or just parse directly from the response body.
	links, err := extractLinks(resp.Body)
	if err != nil {
		return "", nil, err
	}

	return "Fetched content", links, nil
}

func extractLinks(body io.Reader) ([]string, error) {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return links, nil
			}
			return nil, z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
}

func main() {
	fetcher := HttpFetcher{}
	_, urls, err := fetcher.Fetch("https://go.dev/")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found %d URLs\n", len(urls))
	for i, u := range urls {
		if i > 5 {
			fmt.Println("...")
			break
		}
		fmt.Printf("- %s\n", u)
	}
}
