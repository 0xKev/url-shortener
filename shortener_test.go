package shortener_test

import (
	"fmt"
	"strings"
	"testing"

	shortener "github.com/0xKev/url-shortener"
)

const startCounter = 500 // use large initial num to prevent guesses
// use counter with base62 -> reverse to obfusciate generator logic
// tdd top down -> black box -> do not test internal implementation
func TestShortenURL(t *testing.T) {
	shortener := shortener.NewURLShortener()
	t.Run("shorten new urls", func(t *testing.T) {
		shortLink, err := shortener.ShortenURL("google.com")
		assertNoError(t, err)
		assertSuffixLength(t, shortLink, shortener)

		shortLink2, err := shortener.ShortenURL("youtube.com")
		assertNoError(t, err)
		assertNotEqualURL(t, shortLink, shortLink2)

		shortLink3, err := shortener.ShortenURL("github.com")
		assertNoError(t, err)
		assertNotEqualURL(t, shortLink, shortLink3)
		assertNotEqualURL(t, shortLink2, shortLink3)
	})

	t.Run("shorten existing urls", func(t *testing.T) {
		shortLink, _ := shortener.ShortenURL("google.com")
		shortLink2, _ := shortener.ShortenURL("google.com")

		assertEqualURL(t, shortLink, shortLink2)
	})

	t.Run("handle invalid urls", func(t *testing.T) {
		cases := []string{
			"",
			"google",
		}

		for _, c := range cases {
			_, err := shortener.ShortenURL(c)
			if err == nil {
				t.Error("expected an error when shortening invalid URL")
			}
		}
	})

	t.Run("correct suffix length", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			shortLink, _ := shortener.ShortenURL(fmt.Sprintf("example%d.com", i))
			assertSuffixLength(t, shortLink, shortener)
		}
	})

}

func assertSuffixLength(t testing.TB, shortLink string, shortener *shortener.URLShortener) {
	t.Helper()

	shortSuffix := strings.TrimPrefix(shortLink, shortener.Config.Domain())
	if len(shortSuffix) != shortener.Config.URLSuffixLength() {
		t.Fatalf("got %d short suffix length, expected %d", len(shortSuffix), shortener.Config.URLSuffixLength())
	}

}

func assertEqualURL(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got different short links for the same url %v, %v", got, want)
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
