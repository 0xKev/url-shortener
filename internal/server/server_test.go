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
	approvals "github.com/approvals/go-approval-tests"
)

// TODO(HIGH): Refactor server tests using test data create helpers and tables
type StubURLStore struct {
	urlMap        map[string]string
	shortURLCalls []string
	getURLCalls   []string
	urlPair       []model.URLPair
	mu            sync.Mutex
}

func (s *StubURLStore) Save(urlPair model.URLPair) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shortURLCalls = append(s.shortURLCalls, urlPair.ShortSuffix)
	return nil
}

func (s *StubURLStore) Load(shortSuffix string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	baseURL, found := s.urlMap[shortSuffix]
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

// API Tests
func TestAPI_GET_ReturnBaseURL(t *testing.T) {
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
	domain := "https://shortener.com/"
	shortenerServer.SetDomain(domain)

	for idx := range store.urlPair {
		store.urlPair[idx].Domain = domain
	}

	t.Run("returns correct base url via short suffix", func(t *testing.T) {
		getGoogleReq := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)
		getGoogleResponse := httptest.NewRecorder()
		expectedCalls := 2

		shortenerServer.ServeHTTP(getGoogleResponse, getGoogleReq)
		googlePair := testutil.GetURLPairFromResponse(t, getGoogleResponse.Body)

		testutil.AssertStatus(t, getGoogleResponse.Code, http.StatusOK)
		testutil.AssertContentType(t, getGoogleResponse, server.JsonContentType)
		assertURLPairs(t, googlePair, store.urlPair[0])

		getGithubReq := testutil.NewGetAPIExpandedURLRequest(githubShortSuffix)
		getGithubResponse := httptest.NewRecorder()

		shortenerServer.ServeHTTP(getGithubResponse, getGithubReq)
		githubPair := testutil.GetURLPairFromResponse(t, getGithubResponse.Body)

		testutil.AssertStatus(t, getGithubResponse.Code, http.StatusOK)
		testutil.AssertContentType(t, getGithubResponse, server.JsonContentType)
		assertURLPairs(t, githubPair, store.urlPair[1])

		if len(store.getURLCalls) != expectedCalls {
			t.Errorf("expected %d calls to get base url but got %d calls", expectedCalls, len(store.getURLCalls))
		}
	})

	t.Run("returns 404 on missing short links", func(t *testing.T) {
		request := testutil.NewGetAPIExpandedURLRequest(doesNotExistShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		got := response.Code
		want := http.StatusNotFound

		testutil.AssertStatus(t, got, want)
	})
}

func TestAPI_POST_CreateShortURL(t *testing.T) {
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
		// TODO(HIGH): POST REQUESTS SHOULD include the domain in the response as well
		baseUrl := "google.com"
		response := httptest.NewRecorder()
		request := testutil.NewPostAPIShortenURLRequest(baseUrl)
		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)

		urlPair := testutil.GetURLPairFromResponse(t, response.Body)
		testutil.AssertContentType(t, response, server.JsonContentType)
		assertURLPairs(t, urlPair, model.URLPair{ShortSuffix: expectedShortSuffix, BaseURL: baseUrl, Domain: shortenerServer.GetDomain()})

		if len(store.shortURLCalls) != 1 {
			t.Fatalf("got %d calls to shortURLCalls want %d", len(store.shortURLCalls), 1)
		}

		if store.shortURLCalls[0] != expectedShortSuffix {
			t.Errorf("did not store correct url got %q, want %q", store.shortURLCalls[0], expectedShortSuffix)
		}
	})

	t.Run("error on invalid url POST request", func(t *testing.T) {
		response := httptest.NewRecorder()

		request := testutil.NewPostAPIShortenURLRequest(invalidBaseURL) // MAKE THIS RETURN AN ERROR
		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusInternalServerError)
	})
}

func TestAPI_JSONResponse(t *testing.T) {
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
		request := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)

		shortenerServer.ServeHTTP(response, request)

		testutil.AssertNoHTMXRedirect(t, *response.Result())
		testutil.GetURLPairFromResponse(t, response.Body)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
	})

	t.Run("returns valid GET request as JSON", func(t *testing.T) {
		wantedPair := model.URLPair{
			ShortSuffix: googleShortSuffix, BaseURL: "google.com", Domain: shortenerServer.GetDomain(),
		}

		request := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertContentType(t, response, server.JsonContentType)
		testutil.AssertNoHTMXRedirect(t, *response.Result())

		got := testutil.GetURLPairFromResponse(t, response.Body)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		assertURLPairs(t, got, wantedPair)
	})

	t.Run("returns status 200 on valid POST request", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testutil.NewPostAPIShortenURLRequest(store.urlMap[googleShortSuffix])

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertContentType(t, response, server.JsonContentType)
	})

	t.Run("returns valid POST request as JSON", func(t *testing.T) {
		response := httptest.NewRecorder()
		request := testutil.NewPostAPIShortenURLRequest(store.urlMap[googleShortSuffix])

		shortenerServer.ServeHTTP(response, request)

		got := testutil.GetURLPairFromResponse(t, response.Body)

		testutil.AssertContentType(t, response, server.JsonContentType)
		assertURLPairs(t, got, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: store.urlMap[googleShortSuffix], Domain: shortenerServer.GetDomain()})
	})

}

// HTMX
func TestHTMX_Functionality(t *testing.T) {
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

	shortenerServer.SetDomain("https://shortener.domain.com/")

	t.Run("GET /expand/ with HTMX sets HX-Redirect Header", func(t *testing.T) {
		request := testutil.NewGetHTMXExpandedURLRequest(googleShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		testutil.AssertStatus(t, response.Code, http.StatusOK)
		testutil.AssertHTMXRedirect(t, *response.Result(), store.urlMap[googleShortSuffix])
	})

	t.Run("POST /shorten with HTMX returns HTML partial", func(t *testing.T) {
		request := testutil.NewPostHTMXShortenURLRequest(store.urlMap[googleShortSuffix])

		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusOK)

		approvals.VerifyString(t, response.Body.String())
		testutil.AssertContentType(t, response, server.HtmxResponseContentType)

	})
}

// Core Functionality
func TestServer_SetAndRetrieveCorrectDomain(t *testing.T) {
	store := StubURLStore{
		urlMap: map[string]string{
			googleShortSuffix: "google.com",
		},
		urlPair: []model.URLPair{
			{ShortSuffix: googleShortSuffix, BaseURL: "google.com", Domain: ""},
		},
	}

	shortenerServer := server.NewURLShortenerServer(&store, MockURLShortener{})

	t.Run("returns correct domain when setting valid domain", func(t *testing.T) {
		domain := "https://shortener.com/"
		shortenerServer.SetDomain(domain)

		expectedUrlPair := model.URLPair{
			ShortSuffix: googleShortSuffix,
			BaseURL:     store.urlMap[googleShortSuffix],
			Domain:      domain,
		}
		request := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)
		urlPair := testutil.GetURLPairFromResponse(t, response.Body)

		testutil.AssertEqual(t, domain, shortenerServer.GetDomain())
		assertURLPairs(t, urlPair, expectedUrlPair)
	})

	t.Run("returns the default domain when no domain is manually set", func(t *testing.T) {
		expectedUrlPair := model.URLPair{
			ShortSuffix: googleShortSuffix,
			BaseURL:     store.urlMap[googleShortSuffix],
			Domain:      shortenerServer.GetDomain(),
		}

		request := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)
		response := httptest.NewRecorder()

		shortenerServer.ServeHTTP(response, request)

		urlPair := testutil.GetURLPairFromResponse(t, response.Body)

		assertURLPairs(t, urlPair, expectedUrlPair)
	})

}
func TestServer_IndexPage(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	shortenerServer := server.NewURLShortenerServer(&StubURLStore{}, MockURLShortener{})

	shortenerServer.ServeHTTP(response, request)

	testutil.AssertStatus(t, response.Code, http.StatusOK)
}
func TestServer_InvalidRoutes(t *testing.T) {
	shortenerServer := server.NewURLShortenerServer(&StubURLStore{}, MockURLShortener{})
	t.Run("GET request to invalid path returns status 404", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodGet, "/badGet/", nil)

		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
	})

	t.Run("POST request to invalid path returns 404", func(t *testing.T) {
		response := httptest.NewRecorder()
		request, _ := http.NewRequest(http.MethodPost, "/badPost/", nil)

		shortenerServer.ServeHTTP(response, request)

		testutil.AssertStatus(t, response.Code, http.StatusNotFound)
	})
}

// Concurrency
func TestConcurrent_POST_ShortenURL(t *testing.T) {
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
			request := testutil.NewPostAPIShortenURLRequest(store.urlMap[googleShortSuffix])
			shortenerServer.ServeHTTP(response, request)

			testutil.AssertStatus(t, response.Code, http.StatusOK)
			gotPair := testutil.GetURLPairFromResponse(t, response.Body)
			testutil.AssertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, gotPair, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: store.urlMap[googleShortSuffix], Domain: shortenerServer.GetDomain()})
		}()
	}
	wg.Wait()
	if len(store.shortURLCalls) != createCount {
		t.Errorf("expected %d calls to create short url but got %d", createCount, len(store.shortURLCalls))
	}
}

func TestConcurrent_CreateAndGetShortURL(t *testing.T) {
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
			request := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)
			shortenerServer.ServeHTTP(response, request)
			urlPair := testutil.GetURLPairFromResponse(t, response.Body)
			testutil.AssertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, urlPair, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: store.urlMap[googleShortSuffix], Domain: shortenerServer.GetDomain()})

			testutil.AssertStatus(t, response.Code, http.StatusOK)
		}()
	}

	for j := 0; j < createCount; j++ {
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			request := testutil.NewPostAPIShortenURLRequest(store.urlMap[githubShortSuffix])
			shortenerServer.ServeHTTP(response, request)
			urlPair := testutil.GetURLPairFromResponse(t, response.Body)
			testutil.AssertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, urlPair, model.URLPair{ShortSuffix: githubShortSuffix, BaseURL: store.urlMap[githubShortSuffix], Domain: shortenerServer.GetDomain()})

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
func TestConcurrent_GET_ExpandShortURL(t *testing.T) {
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
			request := testutil.NewGetAPIExpandedURLRequest(googleShortSuffix)
			shortenerServer.ServeHTTP(response, request)
			gotPair := testutil.GetURLPairFromResponse(t, response.Body)
			testutil.AssertStatus(t, response.Code, http.StatusOK)
			testutil.AssertContentType(t, response, server.JsonContentType)
			assertURLPairs(t, gotPair, model.URLPair{ShortSuffix: googleShortSuffix, BaseURL: "google.com", Domain: shortenerServer.GetDomain()})
		}()
	}
	wg.Wait()

	if len(store.getURLCalls) != requestCount {
		t.Errorf("expected %d calls to get base url but got %d calls", requestCount, len(store.shortURLCalls))
	}
}

func assertURLPairs(t testing.TB, got, want model.URLPair) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
