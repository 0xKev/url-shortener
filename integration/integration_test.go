package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xKev/url-shortener/internal/base62"
	"github.com/0xKev/url-shortener/internal/server"
	"github.com/0xKev/url-shortener/internal/shortener"
	redis_store "github.com/0xKev/url-shortener/internal/store/redis"
	"github.com/0xKev/url-shortener/internal/testutil"
	"github.com/redis/go-redis/v9"
)

type EncodeFunc func(num uint64) string

func (e EncodeFunc) Encode(num uint64) string {
	return e(num)
}

var encoder shortener.Encoder = EncodeFunc(base62.Encode)

func TestRecordingBaseURLsAndRetrievingThem(t *testing.T) {
	shortenerConfig := shortener.NewDefaultConfig()

	urlShortener := shortener.NewURLShortener(shortenerConfig, encoder)
	storeConfig := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       9, // use only DB 9 for tests
	}
	store, err := redis_store.NewRedisURLStore(storeConfig)
	if err != nil {
		t.Fatalf("error when creating redis store %v", err)
	}
	shortenerServer := server.NewURLShortenerServer(store, urlShortener)

	baseURLs := []string{"google.com", "github.com", "youtube.com"}
	shortLinks := make(map[string]string)

	// Create short URLs
	for _, baseURL := range baseURLs {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewPostShortURLRequest(baseURL))
		testutil.AssertStatus(t, response.Code, http.StatusAccepted)
		shortLinks[baseURL] = response.Body.String()
	}

	// Fetch base URLs from short URLs
	for baseURL, shortSuffix := range shortLinks {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewGetExpandedURLRequest(shortSuffix))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertResponseBody(t, response.Body.String(), baseURL)
	}
}
