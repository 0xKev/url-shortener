package main

import (
	"log"
	"net/http"

	"github.com/0xKev/url-shortener/internal/base62"
	"github.com/0xKev/url-shortener/internal/model"
	"github.com/0xKev/url-shortener/internal/server"
	"github.com/0xKev/url-shortener/internal/shortener"
	redisStore "github.com/0xKev/url-shortener/internal/store/redis"
)

type EncoderFunc func(num uint64) string

func (e EncoderFunc) Encode(num uint64) string {
	return e(num)
}

var encoder shortener.Encoder = EncoderFunc(base62.Encode)

const (
	redisAddr = "localhost:6379"
	redisPass = ""
	redisDB   = 0
)

func main() {
	storeConfig := redisStore.NewRedisConfig(redisAddr, redisPass, redisDB)
	store, err := redisStore.NewRedisURLStore(storeConfig)
	if err != nil {
		log.Fatalf("error when creating redis store %v", err)
	}
	testShortSuffix := "testurl"
	testBaseURL := "https://www.example.com"
	err = store.Save(model.URLPair{testShortSuffix, testBaseURL})
	if err != nil {
		log.Printf("Error saving test URL: %v", err)
	} else {
		log.Printf("Test URL saved: %s -> %s", testShortSuffix, testBaseURL)
	}
	shortenerConfig := shortener.NewDefaultConfig()
	shortenerConfig.SetDomain("localhost:5000")
	shortenerServer := server.NewURLShortenerServer(store, shortener.NewURLShortener(shortenerConfig, encoder))
	log.Fatal(http.ListenAndServe(":5000", shortenerServer))
}
