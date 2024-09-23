package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/0xKev/url-shortener/internal/model"
	server "github.com/0xKev/url-shortener/internal/server"
	testutil "github.com/0xKev/url-shortener/internal/testutil"
)

type StubURLStore struct {
	urlMap        map[string]string
	shortURLCalls []string
	getURLCalls   []string
	urlPair       []model.URLPair
	mu            sync.Mutex
}

func (s *StubURLStore) Save(baseURL, shortLink string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shortURLCalls = append(s.shortURLCalls, baseURL)
	return nil
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
var invalidBaseURL = ""

func TestGETExpandShortURL(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
			githubShortSuffix: "github.com",
		},
		urlPair: []model.URLPair{
			{ShortSuffix: googleShortSuffix, BaseURL: "google.com"},
			{ShortSuffix: githubShortSuffix, BaseURL: "github.com"},
		},
	}

	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{})

	t.Run("returns correct base url via short suffix", func(t *testing.T) {
		getGoogleReq := testutil.NewGetExpandedURLRequest(googleShortSuffix)
		getGoogleResponse := httptest.NewRecorder()
		expectedCalls := 2

		shortenerServer.ServeHTTP(getGoogleResponse, getGoogleReq)
		googlePair := testutil.GetURLPairFromResponse(t, getGoogleResponse.Body)

		testutil.AssertStatus(t, getGoogleResponse.Code, http.StatusOK)
		assertContentType(t, getGoogleResponse, server.JsonContentType)
		assertURLPairs(t, googlePair, store.urlPair[0])

		getGithubReq := testutil.NewGetExpandedURLRequest(githubShortSuffix)
		getGithubResponse := httptest.NewRecorder()

		shortenerServer.ServeHTTP(getGithubResponse, getGithubReq)
		githubPair := testutil.GetURLPairFromResponse(t, getGithubResponse.Body)

		testutil.AssertStatus(t, getGithubResponse.Code, http.StatusOK)
		assertContentType(t, getGithubResponse, server.JsonContentType)
		assertURLPairs(t, githubPair, store.urlPair[1])

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
		[]model.URLPair{},
		sync.Mutex{},
	}
	var expectedShortSuffix = "0000001"
	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		ShortenBaseURLFunc: func(baseURL string) (string, error) {
			if baseURL == invalidBaseURL {
				return "", fmt.Errorf("invalid baseURL %v", baseURL)
			}
			return expectedShortSuffix, nil
		},
	})

	t.Run("records google.com on POST request", func(t *testing.T) {
		baseUrl := "google.com"
		response := httptest.NewRecorder()
		request := testutil.NewPostShortURLRequest(baseUrl)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)

		urlPair := testutil.GetURLPairFromResponse(t, response.Body)
		assertContentType(t, response, server.JsonContentType)
		assertURLPairs(t, urlPair, model.URLPair{ShortSuffix: expectedShortSuffix, BaseURL: baseUrl})

		if len(store.shortURLCalls) != 1 {
			t.Fatalf("got %d calls to shortURLCalls want %d", len(store.shortURLCalls), 1)
		}

		if store.shortURLCalls[0] != expectedShortSuffix {
			t.Errorf("did not store correct url got %q, want %q", store.shortURLCalls[0], expectedShortSuffix)
		}
	})

	t.Run("error on invalid url POST request", func(t *testing.T) {
		response := httptest.NewRecorder()

		request := testutil.NewPostShortURLRequest(invalidBaseURL)
		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusInternalServerError)
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
			gotPair := testutil.GetURLPairFromResponse(t, response.Body)
			testutil.AssertStatus(t, response.Code, http.StatusOK)
			assertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, gotPair, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: "google.com"})
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
	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		ShortenBaseURLFunc: func(baseURL string) (string, error) {
			switch baseURL {
			case "google.com":
				return googleShortSuffix, nil
			case "github.com":
				return githubShortSuffix, nil
			default:
				return "", nil
			}
		},
	})

	for i := 0; i < createCount; i++ {
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			request := testutil.NewPostShortURLRequest(store.urlMap[googleShortSuffix])
			shortenerServer.ServeHTTP(response, request)

			testutil.AssertStatus(t, response.Code, http.StatusOK)
			gotPair := testutil.GetURLPairFromResponse(t, response.Body)
			assertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, gotPair, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: store.urlMap[googleShortSuffix]})
		}()
	}
	wg.Wait()
	if len(store.shortURLCalls) != createCount {
		t.Errorf("expected %d calls to create short url but got %d", createCount, len(store.shortURLCalls))
	}
}

func TestConcurrentCreateAndGetShortURL(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
			githubShortSuffix: "github.com",
		},
	}

	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		ShortenBaseURLFunc: func(baseURL string) (string, error) {
			switch baseURL {
			case "google.com":
				return googleShortSuffix, nil
			case "github.com":
				return githubShortSuffix, nil
			default:
				return "", nil
			}
		},
	})

	requestCount := 1000
	createCount := 2000

	var wg sync.WaitGroup
	wg.Add(requestCount + createCount)

	for i := 0; i < requestCount; i++ {
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			request := testutil.NewGetExpandedURLRequest(googleShortSuffix)
			shortenerServer.ServeHTTP(response, request)
			urlPair := testutil.GetURLPairFromResponse(t, response.Body)
			assertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, urlPair, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: store.urlMap[googleShortSuffix]})

			testutil.AssertStatus(t, response.Code, http.StatusOK)
		}()
	}

	for j := 0; j < createCount; j++ {
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			request := testutil.NewPostShortURLRequest(store.urlMap[githubShortSuffix])
			shortenerServer.ServeHTTP(response, request)
			urlPair := testutil.GetURLPairFromResponse(t, response.Body)
			assertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, urlPair, model.URLPair{ShortSuffix: githubShortSuffix, BaseURL: store.urlMap[githubShortSuffix]})

			testutil.AssertStatus(t, response.Code, http.StatusOK)
		}()
	}
	wg.Wait()

	if len(store.shortURLCalls) != createCount {
		t.Errorf("expected %d calls to create short url but got %d", createCount, len(store.shortURLCalls))
	}

	if len(store.getURLCalls) != requestCount {
		t.Errorf("expected %d calls to get base url but got %d calls", requestCount, len(store.shortURLCalls))
	}
}

func TestInvalidRequestsRoute(t *testing.T) {
	shortenerServer := server.NewURLShortenerServer(&StubURLStore{}, MockURLShortener{})
	t.Run("GET request to invalid path", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/badGet/", nil)

		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("POST request to invalid path", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodPost, "/badPost/", nil)

		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
	})
}

func TestJSONFunctionality(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
			githubShortSuffix: "github.com",
		},
	}
	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{
		ShortenBaseURLFunc: func(baseURL string) (string, error) {
			switch baseURL {
			case "google.com":
				return googleShortSuffix, nil
			case "github.com":
				return githubShortSuffix, nil
			default:
				return "", nil
			}
		},
	})

	t.Run("returns status 200 on valid GET request", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testutil.NewGetExpandedURLRequest(googleShortSuffix)

		shortenerServer.ServeHTTP(response, request)

		testutil.GetURLPairFromResponse(t, response.Body)

		testutil.AssertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns valid GET request as JSON", func(t *testing.T) {
		wantedPair := model.URLPair{
			ShortSuffix: googleShortSuffix, BaseURL: "google.com",
		}

		request := testutil.NewGetExpandedURLRequest(googleShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		assertContentType(t, response, server.JsonContentType)

		got := testutil.GetURLPairFromResponse(t, response.Body)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		assertURLPairs(t, got, wantedPair)
	})

	t.Run("returns status 200 on valid POST request", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testutil.NewPostShortURLRequest(store.urlMap[googleShortSuffix])

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns valid POST request as JSON", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testutil.NewPostShortURLRequest(store.urlMap[googleShortSuffix])

		shortenerServer.ServeHTTP(response, request)

		got := testutil.GetURLPairFromResponse(t, response.Body)

		assertContentType(t, response, server.JsonContentType)
		assertURLPairs(t, got, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: store.urlMap[googleShortSuffix]})
	})

}

func TestIndexPageOKStatus(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	shortenerServer := server.NewURLShortenerServer(&StubURLStore{}, MockURLShortener{})

	shortenerServer.ServeHTTP(response, request)

	testutil.AssertStatus(t, response.Code, http.StatusOK)
}

func assertURLPairs(t testing.TB, got, want model.URLPair) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertContentType(t testing.TB, response *httptest.ResponseRecorder, want string) {
	t.Helper()

	if response.Result().Header.Get("content-type") != want {
		t.Errorf("response did not have content-type of %s, got %v", want, response.Result().Header)
	}
}
