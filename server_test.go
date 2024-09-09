package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	server "github.com/0xKev/url-shortener"
)

type StubURLStore struct {
	urlMap map[string]string
}

func (s *StubURLStore) GetExpandedURL(shortLink string) string {
	return s.urlMap[shortLink]
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
		request := newGetExpandedURLRequest(googleShortPrefix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		assertResponseBody(t, response.Body.String(), store.urlMap[googleShortPrefix])

		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns github.com", func(t *testing.T) {
		request := newGetExpandedURLRequest(githubShortPrefix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		assertResponseBody(t, response.Body.String(), store.urlMap[githubShortPrefix])

		assertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns 404 on missing short links", func(t *testing.T) {
		request := newGetExpandedURLRequest("0000009")
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		got := response.Code
		want := http.StatusNotFound

		assertStatus(t, got, want)
	})
}

func newGetExpandedURLRequest(shortSuffix string) *http.Request {
	request, _ := http.NewRequest("GET", ("/expand/" + shortSuffix), nil)
	return request
}

func assertResponseBody(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
