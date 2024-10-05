package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/0xKev/url-shortener/internal/model"
	"github.com/0xKev/url-shortener/internal/server"
)

func NewGetAPIExpandedURLRequest(shortSuffix string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, server.APIExpandRoute+shortSuffix, nil)
	request.Header.Set("Content-Type", server.JsonContentType)
	return request
}

func NewPostAPIShortenURLRequest(baseURL string) *http.Request {
	urlPair := model.URLPair{BaseURL: baseURL, ShortSuffix: ""}
	body := new(bytes.Buffer)

	err := json.NewEncoder(body).Encode(urlPair)
	if err != nil {
		// log.Fatalf("Error encoding URLPair: %v", err)
		return nil
	}
	request, err := http.NewRequest(http.MethodPost, server.APIShortenRoute, body)
	if err != nil {
		// log.Fatalf("Error creating request: %v", err)
		return nil
	}
	request.Header.Set("Content-Type", server.JsonContentType)
	// log.Printf("Request body: %s", body.String())
	return request
}

func NewGetHTMXExpandedURLRequest(shortSuffix string) *http.Request {
	request, err := http.NewRequest(http.MethodGet, server.HtmxExpandRoute+shortSuffix, nil)
	if err != nil {
		return nil
	}

	request.Header.Set("HX-Request", "true")
	return request
}

func NewPostHTMXShortenURLRequest(baseURL string) *http.Request {
	// use form data instead of json

	formData := url.Values{}
	formData.Set("base-url", baseURL)
	body := strings.NewReader(formData.Encode())

	request, err := http.NewRequest(http.MethodPost, server.HtmxShortenRoute, body)
	if err != nil {
		return nil
	}
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Content-Type", server.HtmxRequestContentType)
	return request
}

func AssertHTMXRedirect(t testing.TB, got http.Response, want string) {
	t.Helper()
	if got.Header.Get("HX-Redirect") != want {
		t.Errorf("expected HX-Redirect to '%v' but got '%v'", want, got.Header.Get("HX-Redirect"))
	}
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

func AssertNoHTMXRedirect(t testing.TB, response http.Response) {
	t.Helper()

	if response.Header.Get("HX-Redirect") != "" {
		t.Fatal("did not expect an HX-Redirect but got one")
	}
}
