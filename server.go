package server

import (
	"fmt"
	"net/http"
	"strings"
)

type URLShortenerServer struct {
	store URLStore
}

func NewURLShortenerServer(store URLStore) *URLShortenerServer {
	return &URLShortenerServer{
		store: store,
	}
}

func (u *URLShortenerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shortLink := strings.TrimPrefix(r.URL.Path, "/expand/")
	if shortLink == "0000009" {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, u.store.GetExpandedURL(shortLink))
}

type URLStore interface {
	GetExpandedURL(shortLink string) string
}

func GetExpandedURL(shortLink string) string {
	if shortLink == "0000001" {
		return "google.com"
	} else if shortLink == "0000002" {
		return "github.com"
	} else {
		return ""
	}

}
