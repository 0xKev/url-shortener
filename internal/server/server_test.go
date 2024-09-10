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

func (s *StubURLStore) GetExpandedURL(shortLink string) string {
	return s.urlMap[shortLink]
}

func (s *StubURLStore) RecordBaseURL(baseURL string) {
	s.shortURLCalls = append(s.shortURLCalls, baseURL)
}

func TestGETExpandShortURL(t *testing.T) {
	var googleShortPrefix = "0000001"
	var githubShortPrefix = "0000002"
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortPrefix: "google.com",
			githubShortPrefix: "github.com",
		},
	}

	shortenerServer := server.NewURLShortenerServer(&store)

	t.Run("returns google.com", func(t *testing.T) {
		request := testutil.NewGetExpandedURLRequest(googleShortPrefix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		t.Log(response.Body.String())
		testutil.AssertResponseBody(t, response.Body.String(), store.urlMap[googleShortPrefix])

		testutil.AssertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns github.com", func(t *testing.T) {
		request := testutil.NewGetExpandedURLRequest(githubShortPrefix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertResponseBody(t, response.Body.String(), store.urlMap[githubShortPrefix])

		testutil.AssertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns 404 on missing short links", func(t *testing.T) {
		request := testutil.NewGetExpandedURLRequest("0000009")
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
	shortenerServer := server.NewURLShortenerServer(&store)

	t.Run("records baseURL on POST", func(t *testing.T) {
		baseUrl := "google.com"
		response := httptest.NewRecorder()
		request := testutil.NewPostShortURLRequest(baseUrl)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusAccepted)

		if len(store.shortURLCalls) != 1 {
			t.Errorf("got %d calls to RecordBaseURL want %d", len(store.shortURLCalls), 1)
		}

		if store.shortURLCalls[0] != baseUrl {
			t.Errorf("did not store correct url got %q, want %q", store.shortURLCalls[0], baseUrl)
		}
	})
}
