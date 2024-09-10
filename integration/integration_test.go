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

	urlShortener := shortener.NewURLShortener(nil, encoder)
	store := memory_store.NewInMemoryURLStore()
	shortenerServer := server.NewURLShortenerServer(store)

	testcases := map[string]string{
		"google.com": domain + "0000000",
		"reddit.com": domain + "0000001",
		"github.com": domain + "0000002",
	}

	for baseURL, _ := range testcases {
		shortURL, _ := urlShortener.ShortenURL(baseURL)
		t.Log(shortURL)
		shortenerServer.ServeHTTP(httptest.NewRecorder(), testutil.NewPostShortURLRequest(baseURL))
	}

	for baseURL, shortURL := range testcases {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, testutil.NewGetExpandedURLRequest(shortURL))
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertResponseBody(t, response.Body.String(), baseURL)
	}
}
