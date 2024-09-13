package server_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	server "github.com/0xKev/url-shortener/internal/server"
	testutil "github.com/0xKev/url-shortener/internal/testutil"
)

type StubURLStore struct {
	urlMap        map[string]string
	shortURLCalls []string
	getURLCalls   []string
	mu            sync.Mutex
}

func (s *StubURLStore) Save(baseURL, shortLink string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shortURLCalls = append(s.shortURLCalls, baseURL)
}

func (s *StubURLStore) Load(shortLink string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	baseURL, found := s.urlMap[shortLink]
	s.getURLCalls = append(s.getURLCalls, baseURL)
	return baseURL, found
}

type MockURLShortener struct {
	ShortenBaseURLFunc func(baseURL string) (string, error)
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

	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{})

	t.Run("returns correct base url via short suffix", func(t *testing.T) {
		getGoogleReq := testutil.NewGetExpandedURLRequest(googleShortSuffix)
		getGoogleResponse := httptest.NewRecorder()
		expectedCalls := 2

		shortenerServer.ServeHTTP(getGoogleResponse, getGoogleReq)
		testutil.AssertStatus(t, getGoogleResponse.Code, http.StatusOK)
		testutil.AssertResponseBody(t, getGoogleResponse.Body.String(), store.urlMap[googleShortSuffix])

		getGithubReq := testutil.NewGetExpandedURLRequest(githubShortSuffix)
		getGithubResponse := httptest.NewRecorder()

		shortenerServer.ServeHTTP(getGithubResponse, getGithubReq)
		testutil.AssertStatus(t, getGithubResponse.Code, http.StatusOK)
		testutil.AssertResponseBody(t, getGithubResponse.Body.String(), store.urlMap[githubShortSuffix])

		if len(store.getURLCalls) != expectedCalls {
			t.Errorf("expected %d calls to get base url but got %d calls", expectedCalls, len(store.getURLCalls))
		}
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
		nil,
		sync.Mutex{},
	}
	var expectedShortSuffix = "0000001"
	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		ShortenBaseURLFunc: func(baseURL string) (string, error) {
			return expectedShortSuffix, nil
		},
	})

	t.Run("records google.com on POST request", func(t *testing.T) {
		baseUrl := "google.com"
		response := httptest.NewRecorder()
		request := testutil.NewPostShortURLRequest(baseUrl)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusAccepted)

		if len(store.shortURLCalls) != 1 {
			t.Fatalf("got %d calls to shortURLCalls want %d", len(store.shortURLCalls), 1)
		}

		if store.shortURLCalls[0] != expectedShortSuffix {
			t.Errorf("did not store correct url got %q, want %q", store.shortURLCalls[0], expectedShortSuffix)
		}
	})
}

func TestConcurrentGETExpandShortURL(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
			githubShortSuffix: "github.com",
		},
	}

	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{})
	requestCount := 1000

	var wg sync.WaitGroup
	wg.Add(requestCount)

	for i := 0; i < requestCount; i++ {
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			request := testutil.NewGetExpandedURLRequest(googleShortSuffix)
			shortenerServer.ServeHTTP(response, request)
			testutil.AssertStatus(t, response.Code, http.StatusOK)
			testutil.AssertResponseBody(t, response.Body.String(), store.urlMap[googleShortSuffix])
		}()
	}
	wg.Wait()

	if len(store.getURLCalls) != requestCount {
		t.Errorf("expected %d calls to get base url but got %d calls", requestCount, len(store.shortURLCalls))
	}
}

func TestConcurrentCreateShortURL(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
			githubShortSuffix: "github.com",
		},
	}

	createCount := 1000
	var wg sync.WaitGroup
	wg.Add(createCount)

	for i := 0; i < createCount; i++ {
		go func() {
			defer wg.Done()
			shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{})
			response := httptest.NewRecorder()
			request := testutil.NewPostShortURLRequest(store.urlMap[googleShortSuffix])
			shortenerServer.ServeHTTP(response, request)

			testutil.AssertStatus(t, response.Code, http.StatusAccepted)
		}()
	}
	wg.Wait()
	if len(store.shortURLCalls) != createCount {
		t.Errorf("expected %d calls to create short url but got %d", createCount, len(store.shortURLCalls))
	}
}
