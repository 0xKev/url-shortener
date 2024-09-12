package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xKev/url-shortener/internal/base62"
	"github.com/0xKev/url-shortener/internal/server"
	"github.com/0xKev/url-shortener/internal/shortener"
	memory_store "github.com/0xKev/url-shortener/internal/store"
	"github.com/0xKev/url-shortener/internal/testutil"
)

type EncodeFunc func(num uint64) string

func (e EncodeFunc) Encode(num uint64) string {
	return e(num)
}

var encoder shortener.Encoder = EncodeFunc(base62.Encode)
var domain = "s.nykevin.com/"

func TestRecordingBaseURLsAndRetrievingThem(t *testing.T) {

	shortenerConfig := shortener.NewDefaultConfig()
	shortenerConfig.SetDomain(domain)
	shortenerConfig.SetURLCounter(0)

	urlShortener := shortener.NewURLShortener(shortenerConfig, encoder)
	store := memory_store.NewInMemoryURLStore()
	shortenerServer := server.NewURLShortenerServer(store, urlShortener)

	testcases := map[string]string{
		"google.com": "0000001",
		"reddit.com": "0000002",
		"github.com": "0000003",
	}

	for baseURL, shortLink := range testcases {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewPostShortURLRequest(baseURL))
		testutil.AssertStatus(t, response.Code, http.StatusAccepted)
		testutil.AssertResponseBody(t, response.Body.String(), shortLink)
	}

	for baseURL, shortSuffix := range testcases {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewGetExpandedURLRequest(shortSuffix))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertResponseBody(t, response.Body.String(), baseURL)
	}
}
