package main

import (
	"log"
	"net/http"

	server "github.com/0xKev/url-shortener"
	store "github.com/0xKev/url-shortener/internal/store"
)

func main() {
	shortenerServer := server.NewURLShortenerServer(store.NewInMemoryURLStore())
	log.Fatal(http.ListenAndServe(":5000", shortenerServer))
}
