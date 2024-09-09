package main

import (
	"log"
	"net/http"

	server "github.com/0xKev/url-shortener"
)

type InMemoryURLStore struct{}

func (i *InMemoryURLStore) GetExpandedURL(shortLink string) string {
	return "google.com"
}

func main() {
	shortenerServer := server.NewURLShortenerServer(&InMemoryURLStore{})
	log.Fatal(http.ListenAndServe(":5000", shortenerServer))
}
