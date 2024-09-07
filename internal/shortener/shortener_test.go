package shortener_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	shortener "github.com/0xKev/url-shortener/internal/shortener"
)

const (
	startCounter = 500
	google       = "google.com"
	youtube      = "youtube.com"
	github       = "github.com"
	reddit       = "reddit.com"
) // use large initial num to prevent guesses
// use counter with base62 -> reverse to obfusciate generator logic
// tdd top down -> black box -> do not test internal implementation

func TestShortenURL(t *testing.T) {
	urlShortener := setUpShortener()
	t.Run("shorten new urls", func(t *testing.T) {
		shortLink, err := urlShortener.ShortenURL(google)
		assertNoError(t, err)
		assertSuffixLength(t, shortLink, urlShortener)

		shortLink2, err := urlShortener.ShortenURL(youtube)
		assertNoError(t, err)
		assertNotEqualURL(t, shortLink, shortLink2)

		shortLink3, err := urlShortener.ShortenURL(github)
		assertNoError(t, err)
		assertNotEqualURL(t, shortLink, shortLink3)
		assertNotEqualURL(t, shortLink2, shortLink3)
	})

	t.Run("shorten existing urls", func(t *testing.T) {
		shortLink, _ := urlShortener.ShortenURL(google)
		shortLink2, _ := urlShortener.ShortenURL(google)

		assertEqualURL(t, shortLink, shortLink2)
	})

	t.Run("handle invalid urls", func(t *testing.T) {
		cases := []struct {
			baseURL     string
			expectedErr shortener.InvalidURLError
		}{
			{"", shortener.InvalidURLError{shortener.ErrEmptyURL, ""}},
			{"google", shortener.InvalidURLError{shortener.ErrNoDomainURL, "google"}},
		}

		for _, c := range cases {
			_, err := urlShortener.ShortenURL(c.baseURL)
			assertError(t, err)
			if !errors.Is(err, c.expectedErr) {
				t.Fatalf("expected error %v, but got %v", c.expectedErr.Error(), err)
			}
		}
	})

	t.Run("correct suffix length", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			shortLink, _ := urlShortener.ShortenURL(fmt.Sprintf("example%d.com", i))
			assertSuffixLength(t, shortLink, urlShortener)
		}
	})

	t.Run("expect error when URLCounter is over the max limit", func(t *testing.T) {
		config := shortener.NewDefaultConfig()
		config.SetURLCounter(config.URLCounterLimit() - 2) // num of valid cases
		urlShortener := shortener.NewURLShortener(config)
		counterLimit := urlShortener.Config.URLCounterLimit()
		wantError := shortener.ExceedCounterError{counterLimit, counterLimit}

		cases := []struct {
			baseURL       string
			expectedError error
		}{
			{google, nil},
			{github, nil},
			{youtube, wantError},
			{reddit, wantError},
		}

		for _, c := range cases {
			_, err := urlShortener.ShortenURL(c.baseURL)
			t.Logf("Current counter: %d", config.URLCounter())
			if c.expectedError != nil {
				assertError(t, err)

				if !errors.As(err, &wantError) {
					t.Fatalf("expected error %v but did not get one", c.expectedError)
				}
			}
		}
	})
}

func TestExpandURL(t *testing.T) {
	urlShortener := setUpShortener()
	t.Run("shortened url should return original url", func(t *testing.T) {
		shortLink, err := urlShortener.ShortenURL(google)
		assertNoError(t, err)
		originalLink, _ := urlShortener.ExpandURL(shortLink)
		assertEqualURL(t, google, originalLink)

		shortLink2, err := urlShortener.ShortenURL(originalLink)
		assertNoError(t, err)

		assertEqualURL(t, shortLink, shortLink2)
	})

	t.Run("expect error when expanding non existent URLs", func(t *testing.T) {
		_, err := urlShortener.ExpandURL(google)
		expectedError := shortener.ShortURLNotFoundError{ShortURL: google}
		assertError(t, err)
		if !errors.As(err, &expectedError) {
			t.Fatalf("expected error %v but got none", expectedError.Error())
		}
	})

}

func TestConcurrentOperations(t *testing.T) {
	urlShortener := setUpShortener()

	cases := []string{
		youtube,
		google,
		github,
	}

	results := make(map[string]string)

	var wg sync.WaitGroup

	for _, url := range cases {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			shortLink, err := urlShortener.ShortenURL(url)
			assertNoError(t, err)
			results[url] = shortLink
			assertSuffixLength(t, shortLink, urlShortener)

			shortLink2, err := urlShortener.ShortenURL("example.com")
			assertNoError(t, err)
			assertSuffixLength(t, shortLink, urlShortener)
			assertNotEqualURL(t, shortLink, shortLink2)
		}(url)
	}

	wg.Wait()

	assertEqual(t, len(cases), len(results))
	shortLink, err := urlShortener.ShortenURL(cases[0])
	expandedLink, _ := urlShortener.ExpandURL(shortLink)
	assertEqualURL(t, expandedLink, cases[0])
	assertNoError(t, err)

	for originalURL, shortURL := range results {
		shortLink, _ := urlShortener.ShortenURL(originalURL)
		assertEqualURL(t, shortLink, shortURL)
		expandedLink, _ := urlShortener.ExpandURL(shortLink)
		assertEqualURL(t, originalURL, expandedLink)
	}
}

func setUpShortener() *shortener.URLShortener {
	defaultConfig := shortener.NewDefaultConfig()
	urlShortener := shortener.NewURLShortener(defaultConfig) // nil loads in default config
	return urlShortener
}

func assertSuffixLength(t testing.TB, shortLink string, shortener *shortener.URLShortener) {
	t.Helper()

	shortSuffix := strings.TrimPrefix(shortLink, shortener.Config.Domain())
	if len(shortSuffix) != int(shortener.Config.URLSuffixLength()) {
		t.Fatalf("got %d short suffix length, expected %d", len(shortSuffix), shortener.Config.URLSuffixLength())
	}
}

func assertEqual[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("expected %v but got %v", want, got)
	}
}

func assertError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error but got %v", err)
	}
}

func assertEqualURL(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got different short links for the same url : %v, %v", got, want)
	}
}

func assertNotEqualURL(t testing.TB, got, want string) {
	t.Helper()

	if got == want {
		t.Errorf("expected different short links for different url 	%v, %v", got, want)
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("should not have gotten an error but got error %q", err)
	}
}
