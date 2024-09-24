package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/0xKev/url-shortener/internal/model"
	"github.com/0xKev/url-shortener/internal/server"
)

func NewGetExpandedURLRequest(shortSuffix string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, server.ExpandRoute+shortSuffix, nil)
	return request
}

func NewPostShortURLRequest(baseURL string) *http.Request {
	formData := url.Values{}
	formData.Set("base-url", baseURL)
	body := strings.NewReader(formData.Encode())
	request, _ := http.NewRequest(http.MethodPost, server.ShortenRoute, body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

func AssertNoError(t testing.TB, err error) {
	if err != nil {
		t.Errorf("expected no error but got %v", err)
	}
}

func AssertError(t testing.TB, err error) {
	if err == nil {
		t.Error("expected an error but got none")
	}
}

func GetURLPairFromResponse(t testing.TB, body io.Reader) model.URLPair {
	t.Helper()

	got := model.URLPair{}

	err := json.NewDecoder(body).Decode(&got)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of URLPair, '%v'", body, err)
	}
	return got
}
