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

var encoder shortener.Encoder = EncodeFunc(base62.Encode)

const (
	redisAddr = "localhost:6379"
	redisPass = ""
	redisDB   = 9 // use 9 for tests
)

func TestRecordingBaseURLsAndRetrievingThem(t *testing.T) {
	shortenerConfig := shortener.NewDefaultConfig()

	urlShortener := shortener.NewURLShortener(shortenerConfig, encoder)
	storeConfig := redis_store.NewRedisConfig(redisAddr, redisPass, redisDB)

	store, err := redis_store.NewRedisURLStore(storeConfig)
	if err != nil {
		t.Fatalf("error when creating redis store %v", err)
	}
	shortenerServer := server.NewURLShortenerServer(store, urlShortener)

	baseURLs := []string{"google.com", "github.com", "youtube.com"}
	expectedPairs := []model.URLPair{}

	// Create short URLs
	for _, baseURL := range baseURLs {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewPostShortURLRequest(baseURL))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		gotPair := testutil.GetURLPairFromResponse(t, response.Body)
		expectedPairs = append(expectedPairs, gotPair)
	}

	// Fetch base URLs from short URLs
	gotPairs := []model.URLPair{}
	for _, urlPair := range expectedPairs {
		response := httptest.NewRecorder()
		request := testutil.NewGetExpandedURLRequest(urlPair.ShortSuffix)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		gotPairs = append(gotPairs, testutil.GetURLPairFromResponse(t, response.Body))
	}

	if !reflect.DeepEqual(expectedPairs, gotPairs) {
		t.Fatalf("getting base url after creating short url does not return the same url")
	}

	response := httptest.NewRecorder()
	shortenerServer.ServeHTTP(response, testutil.NewPostShortURLRequest(""))
	testutil.AssertStatus(t, response.Code, http.StatusInternalServerError) // change to status 404 instead of 500
}
