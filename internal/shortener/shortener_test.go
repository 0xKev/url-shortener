package shortener_test

import (
	"errors"
	"fmt"
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

type MockEncoder struct {
	encodeCalls []uint64
	encodeFunc  func(uint64) string
}

func (m *MockEncoder) Encode(num uint64) string {
	m.encodeCalls = append(m.encodeCalls, num)
	return m.encodeFunc(num)
}

func TestShortenURL(t *testing.T) {
	t.Run("shorten new urls", func(t *testing.T) {
		urlShortener, _ := setUpShortener()

		shortLink, err := urlShortener.ShortenURL(google)
		assertNoError(t, err)
		assertSuffixLength(t, shortLink, urlShortener)

		shortLink2, err := urlShortener.ShortenURL(youtube)
		assertNoError(t, err)
		assertSuffixLength(t, shortLink2, urlShortener)

		shortLink3, err := urlShortener.ShortenURL(github)
		assertNoError(t, err)
		assertSuffixLength(t, shortLink3, urlShortener)
	})

	t.Run("counter correctly increments when shortening url ", func(t *testing.T) {
		urlShortener, encoder := setUpShortener()
		urlShortener.ShortenURL(google)
		assertEqual(t, urlShortener.Config.URLCounter(), encoder.encodeCalls[0])

		urlShortener.ShortenURL(google)
		assertEqual(t, urlShortener.Config.URLCounter(), encoder.encodeCalls[1])

		assertEqual(t, len(encoder.encodeCalls), 2)

	})

	t.Run("shorten same base url results in new short links", func(t *testing.T) {
		urlShortener, _ := setUpShortener()

		shortLink, _ := urlShortener.ShortenURL(google)
		shortLink2, _ := urlShortener.ShortenURL(google)

		assertNotEqualURL(t, shortLink, shortLink2)
	})

	t.Run("handle invalid urls", func(t *testing.T) {
		urlShortener, _ := setUpShortener()

		cases := []struct {
			baseURL     string
			expectedErr shortener.InvalidURLError
		}{
			{"", shortener.InvalidURLError{shortener.ErrEmptyURL, ""}},
			{"google", shortener.InvalidURLError{shortener.ErrNoDomainURL, "google"}},
			{"bad-base-url", shortener.InvalidURLError{shortener.ErrNoDomainURL, "bad-base-url"}},
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
		urlShortener, _ := setUpShortener()

		for i := 0; i < 1000; i++ {
			shortLink, _ := urlShortener.ShortenURL(fmt.Sprintf("example%d.com", i))
			assertSuffixLength(t, shortLink, urlShortener)
		}
	})

	t.Run("expect error when URLCounter is over the max limit", func(t *testing.T) {
		config := shortener.NewDefaultConfig()
		config.SetURLCounter(config.URLCounterLimit() - 2) // num of valid cases
		encoder := MockEncoder{
			encodeFunc: func(num uint64) string {
				return fmt.Sprintf("%07d", num)
			},
		}
		urlShortener := shortener.NewURLShortener(config, &encoder)
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

func TestNewURLShortener(t *testing.T) {
	config := shortener.NewDefaultConfig()
	encoder := MockEncoder{}

	urlShortener := shortener.NewURLShortener(config, &encoder)

	assertEqual(t, urlShortener.Config, config)
}

func TestConcurrentShortening(t *testing.T) {
	urlShortener, _ := setUpShortener()

	for i := 0; i < 1000; i++ {
		go func(baseURL string) {
			shortLink, _ := urlShortener.ShortenURL(baseURL)
			assertSuffixLength(t, shortLink, urlShortener)
		}(fmt.Sprintf("example%d.com", i))
	}
}

func BenchmarkValidShortening(b *testing.B) {
	urlShortener, _ := setUpShortener()
	b.ResetTimer()
	for i := 0; i < 1000; i++ {
		shortLink, _ := urlShortener.ShortenURL(fmt.Sprintf("example%d.com", i))
		assertSuffixLength(b, shortLink, urlShortener)
	}
}

func setUpShortener() (*shortener.URLShortener, *MockEncoder) {
	defaultConfig := shortener.NewDefaultConfig()
	mockEncoder := MockEncoder{
		encodeFunc: func(num uint64) string {
			return fmt.Sprintf("%07d", num)
		},
	}
	urlShortener := shortener.NewURLShortener(defaultConfig, &mockEncoder) // nil loads in default config
	return urlShortener, &mockEncoder
}

func assertSuffixLength(t testing.TB, shortSuffix string, shortener *shortener.URLShortener) {
	t.Helper()

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
