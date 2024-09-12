package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	server "github.com/0xKev/url-shortener/internal/server"
	testutil "github.com/0xKev/url-shortener/internal/testutil"
)

type StubURLStore struct {
	urlMap        map[string]string
	shortURLCalls []string
}

func (s *StubURLStore) Save(baseURL, shortLink string) {
	s.shortURLCalls = append(s.shortURLCalls, baseURL)
}

func (s *StubURLStore) Load(shortLink string) (string, bool) {
	baseURL, found := s.urlMap[shortLink]
	return baseURL, found
}

type MockURLShortener struct {
	GetExpandedURLFunc func(shortLink string) string
	ShortenBaseURLFunc func(baseURL string) (string, error)
}

func (m MockURLShortener) GetExpandedURL(shortLink string) string {
	if m.GetExpandedURLFunc != nil {
		return m.GetExpandedURLFunc(shortLink)
	}
	return ""
}

func (m MockURLShortener) ShortenURL(baseURL string) (string, error) {
	if m.ShortenBaseURLFunc != nil {
		return m.ShortenBaseURLFunc(baseURL)
	}
	return "", nil
}

var googleShortSuffix = "0000001"
var githubShortSuffix = "0000002"
var doesNotExistShortSuffix = "1000000"

func TestGETExpandShortURL(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
			githubShortSuffix: "github.com",
		},
	}

	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		GetExpandedURLFunc: func(shortLink string) string {
			switch shortLink {
			case googleShortSuffix:
				return store.urlMap[googleShortSuffix]
			case githubShortSuffix:
				return store.urlMap[githubShortSuffix]
			default:
				return ""
			}
		},
	})

	t.Run("returns google.com via short suffix", func(t *testing.T) {
		request := testutil.NewGetExpandedURLRequest(googleShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertResponseBody(t, response.Body.String(), store.urlMap[googleShortSuffix])
	})

	t.Run("returns github.com via short suffix", func(t *testing.T) {
		request := testutil.NewGetExpandedURLRequest(githubShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertResponseBody(t, response.Body.String(), store.urlMap[githubShortSuffix])
	})

	t.Run("returns 404 on missing short links", func(t *testing.T) {
		request := testutil.NewGetExpandedURLRequest(doesNotExistShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		got := response.Code
		want := http.StatusNotFound

		testutil.AssertStatus(t, got, want)
	})
}

func TestCreateShortURL(t *testing.T) {
	store := StubURLStore{
		map[string]string{},
		nil,
	}
	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		ShortenBaseURLFunc: func(baseURL string) (string, error) {
			return "0000001", nil
		},
	})

	t.Run("records baseURL on POST", func(t *testing.T) {
		baseUrl := "google.com"
		response := httptest.NewRecorder()
		request := testutil.NewPostShortURLRequest(baseUrl)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusAccepted)

		if len(store.shortURLCalls) != 1 {
			t.Fatalf("got %d calls to shortURLCalls want %d", len(store.shortURLCalls), 1)
		}

		if store.shortURLCalls[0] != baseUrl {
			t.Errorf("did not store correct url got %q, want %q", store.shortURLCalls[0], baseUrl)
		}
	})
}
