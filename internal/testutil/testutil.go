package testutil

import (
	"net/http"
	"testing"
)

func NewGetExpandedURLRequest(shortSuffix string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/expand/"+shortSuffix, nil)
	return request
}

func NewPostShortURLRequest(baseURL string) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, "/shorten/"+baseURL, nil)
	return request
}

func AssertResponseBody(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("response body is wrong, got %q want %q", got, want)
	}
}

func AssertStatus(t testing.TB, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
