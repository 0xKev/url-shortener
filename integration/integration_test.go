package integration

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/0xKev/url-shortener/internal/base62"
	"github.com/0xKev/url-shortener/internal/model"
	"github.com/0xKev/url-shortener/internal/server"
	"github.com/0xKev/url-shortener/internal/shortener"
	redis_store "github.com/0xKev/url-shortener/internal/store/redis"
	"github.com/0xKev/url-shortener/internal/testutil"
)

type EncodeFunc func(num uint64) string

func (e EncodeFunc) Encode(num uint64) string {
	return e(num)
}

var (
	encoder  shortener.Encoder = EncodeFunc(base62.Encode)
	baseURLs                   = []string{"google.com", "github.com", "youtube.com"}
)

const (
	redisAddr = "localhost:6379"
	redisPass = ""
	redisDB   = 9 // use 9 for tests
)

func TestAPIRecordingBaseURLsAndRetrievingThem(t *testing.T) {
	shortenerConfig := shortener.NewDefaultConfig()

	urlShortener := shortener.NewURLShortener(shortenerConfig, encoder)
	storeConfig := redis_store.NewRedisConfig(redisAddr, redisPass, redisDB)

	store, err := redis_store.NewRedisURLStore(storeConfig)
	if err != nil {
		t.Fatalf("error when creating redis store %v", err)
	}
	shortenerServer := server.NewURLShortenerServer(store, urlShortener)

	expectedPairs := []model.URLPair{}

	// Create short URLs
	for _, baseURL := range baseURLs {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewPostAPIShortenURLRequest(baseURL))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertContentType(t, response, server.JsonContentType)
		gotPair := testutil.GetURLPairFromResponse(t, response.Body)
		expectedPairs = append(expectedPairs, gotPair)
	}

	// Fetch base URLs from short URLs
	gotPairs := []model.URLPair{}
	for _, urlPair := range expectedPairs {
		response := httptest.NewRecorder()
		request := testutil.NewGetAPIExpandedURLRequest(urlPair.ShortSuffix)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertContentType(t, response, server.JsonContentType)
		gotPairs = append(gotPairs, testutil.GetURLPairFromResponse(t, response.Body))
	}

	if !reflect.DeepEqual(expectedPairs, gotPairs) {
		t.Fatalf("getting base url after creating short url does not return the same url")
	}

	response := httptest.NewRecorder()
	shortenerServer.ServeHTTP(response, testutil.NewPostAPIShortenURLRequest(""))
	testutil.AssertStatus(t, response.Code, http.StatusInternalServerError) // change to status 404 instead of 500
}

func TestHTMXRecordingBaseURLsAndRetrievingThem(t *testing.T) {
	shortenerConfig := shortener.NewDefaultConfig()
	shortSuffixes := fetchShortSuffixes(t, shortenerConfig.URLCounter(), uint64(len(baseURLs)), encoder)
	t.Log("length of shortSuffixes is", shortSuffixes)
	urlShortener := shortener.NewURLShortener(shortenerConfig, encoder)
	storeConfig := redis_store.NewRedisConfig(redisAddr, redisPass, redisDB)

	store, err := redis_store.NewRedisURLStore(storeConfig)
	if err != nil {
		t.Fatalf("error creating redis store %v", err)
	}

	shortenerServer := server.NewURLShortenerServer(store, urlShortener)

	shortToBaseURL := make(map[string]string)

	// Create short URLs
	for idx, baseURL := range baseURLs {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewPostHTMXShortenURLRequest(baseURL))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertContentType(t, response, server.HtmxResponseContentType)

		shortToBaseURL[shortSuffixes[idx]] = baseURL
	}

	// Fetch base URLs from short URLS

	for _, shortSuffix := range shortSuffixes {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewGetHTMXExpandedURLRequest(shortSuffix))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertContentType(t, response, server.HtmxResponseContentType)
		baseURL := shortToBaseURL[shortSuffix]
		testutil.AssertHTMXRedirect(t, *response.Result(), baseURL)
	}

}

func fetchShortSuffixes(t testing.TB, start uint64, increments uint64) []string {
	t.Helper()
	var shortSuffixes []string

	for range increments {
		start += 1
		shortSuffixes = append(shortSuffixes, string(base62.Encode(start)))
	}

	return shortSuffixes
}
