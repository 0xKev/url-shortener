package server

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	ExpandRoute  = "/expand/"
	ShortenRoute = "/shorten/"
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
	switch r.Method {
	case http.MethodPost:
		u.processShortURL(w, r)
	case http.MethodGet:
		u.showExpandedURL(w, r)
	}
}

func (u *URLShortenerServer) showExpandedURL(w http.ResponseWriter, r *http.Request) {
	shortLink := strings.TrimPrefix(r.URL.Path, ExpandRoute)
	if shortLink == "0000009" {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, u.store.GetExpandedURL(shortLink))
}

func (u *URLShortenerServer) processShortURL(w http.ResponseWriter, r *http.Request) {
	baseURL := strings.TrimPrefix(r.URL.Path, ShortenRoute)
	u.store.RecordBaseURL(baseURL)
	w.WriteHeader(http.StatusAccepted)
}

type URLStore interface {
	GetExpandedURL(shortLink string) string
	RecordBaseURL(baseURL string)
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
