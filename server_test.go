package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	server "github.com/0xKev/url-shortener"
)

func TestGETExpandShortURL(t *testing.T) {
	var google = "0000001"
	var github = "0000002"
	t.Run("returns google.com", func(t *testing.T) {
		request := newGetExpandedURLRequest(google)
		response := httptest.NewRecorder()

		server.URLShortenerServer(response, request)

		got := response.Body.String()
		want := "google.com"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("returns github.com", func(t *testing.T) {
		request := newGetExpandedURLRequest(github)
		response := httptest.NewRecorder()

		server.URLShortenerServer(response, request)

		got := response.Body.String()
		want := "github.com"

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}

func newGetExpandedURLRequest(shortSuffix string) *http.Request {
	request, _ := http.NewRequest("GET", ("/expand/" + shortSuffix), nil)
	return request
}
