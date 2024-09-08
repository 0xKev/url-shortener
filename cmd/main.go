package main

import (
	"log"
	"net/http"

	server "github.com/0xKev/url-shortener"
)

func main() {
	handler := http.HandlerFunc(server.URLShortenerServer)
	log.Fatal(http.ListenAndServe(":5000", handler))
}
