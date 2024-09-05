package shortener_test

import (
	"strings"
	"testing"

	shortener "github.com/0xKev/url-shortener"
)

const startCounter = 500 // use large initial num to prevent guesses
// use counter with base62
// tdd top down -> black box -> do not test internal implementation
func TestShortenURL(t *testing.T) {
	t.Run("shorten new urls", func(t *testing.T) {
		shortLink, err := shortener.ShortenURL("google.com")
		assertNoError(t, err)
		assertSuffixLength(t, shortLink)

		shortLink2, err := shortener.ShortenURL("youtube.com")
		assertNoError(t, err)
		assertNotEqualURL(t, shortLink, shortLink2)

		shortLink3, err := shortener.ShortenURL("github.com")
		assertNoError(t, err)
		assertNotEqualURL(t, shortLink, shortLink3)
		assertNotEqualURL(t, shortLink2, shortLink3)
	})

	t.Run("shorten existing urls", func(t *testing.T) {
		shortLink, err := shortener.ShortenURL("google.com")
		assertNoError(t, err)

		shortLink2, err := shortener.ShortenURL("google.com")
		assertNoError(t, err)

		assertEqualURL(t, shortLink, shortLink2)
	})

	t.Run("handle invalid urls", func(t *testing.T) {
		_, err := shortener.ShortenURL("")
		if err == nil {
			t.Error("expected an error when shortening invalid URL")
		}

	})
}

func assertSuffixLength(t testing.TB, shortLink string) {
	t.Helper()

	shortSuffix := strings.TrimPrefix(shortLink, shortener.Domain)
	if len(shortSuffix) != shortener.UrlSuffixLength {
		t.Errorf("got %d short suffix length, expected %d", len(shortSuffix), shortener.UrlSuffixLength)
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
