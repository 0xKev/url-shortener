package main

import (
	"log"
	"net/http"

	server "github.com/0xKev/url-shortener/internal/server"
	store "github.com/0xKev/url-shortener/internal/store"
)

func main() {
	shortenerServer := server.NewURLShortenerServer(store.NewInMemoryURLStore())
	log.Fatal(http.ListenAndServe(":5000", shortenerServer))
}
