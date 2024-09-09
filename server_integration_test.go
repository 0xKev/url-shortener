package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	server "github.com/0xKev/url-shortener"
	memory_store "github.com/0xKev/url-shortener/internal/store"
)

func TestRecordingBaseURLsAndRetrievingThem(t *testing.T) {
	store := memory_store.NewInMemoryURLStore()
	shortenerServer := server.NewURLShortenerServer(store)

	urls := []string{
		"google.com",
		"reddit.com",
		"github.com",
	}
	for _, url := range urls {
		shortenerServer.ServeHTTP(httptest.NewRecorder(), newPostShortURLRequest(url))
	}

	for _, url := range urls {
		response := httptest.NewRecorder()
		shortenerServer.ServeHTTP(response, newGetExpandedURLRequest(url))
		assertStatus(t, response.Code, http.StatusOK)
		assertResponseBody(t, response.Body.String(), url)
	}
}
