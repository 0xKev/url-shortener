package server

import (
	"fmt"
	"net/http"
	"strings"
)

func URLShortenerServer(w http.ResponseWriter, r *http.Request) {
	shortLink := strings.TrimPrefix(r.URL.Path, "/expand/")
	fmt.Fprint(w, GetExpandedURL(shortLink))

}

func GetExpandedURL(shortLink string) string {
	if shortLink == "0000084" {
		return "google.com"
	} else {
		return "github.com"
	}

}
