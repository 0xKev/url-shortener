package main

import (
	"log"
	"net/http"

	"github.com/0xKev/url-shortener/internal/base62"
	"github.com/0xKev/url-shortener/internal/server"
	"github.com/0xKev/url-shortener/internal/shortener"
	store "github.com/0xKev/url-shortener/internal/store/memory"
)

type EncoderFunc func(num uint64) string

func (e EncoderFunc) Encode(num uint64) string {
	return e(num)
}

var encoder shortener.Encoder = EncoderFunc(base62.Encode)

func main() {
	shortenerServer := server.NewURLShortenerServer(store.NewInMemoryURLStore(), shortener.NewURLShortener(nil, encoder))
	log.Fatal(http.ListenAndServe(":5000", shortenerServer))
}
