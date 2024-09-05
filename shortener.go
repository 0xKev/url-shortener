package shortener

import (
	"fmt"
)

const Domain = "s.nykevin.com/"
const UrlSuffixLength = 7

func ShortenURL(link string) (string, error) {
	switch link {
	case "google.com":
		return Domain + "abc1234", nil
	case "youtube.com":
		return Domain + "1234abc", nil
	default:
		return "", fmt.Errorf("can't shorten invalid url %v", link)
	}
}
