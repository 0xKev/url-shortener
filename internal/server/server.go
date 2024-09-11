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

type URLShortener interface {
	GetExpandedURL(shortLink string) string
	ShortenBaseURL(baseURL string) string
}

type URLShortenerServer struct {
	store     URLStore
	shortener URLShortener
}

func NewURLShortenerServer(store URLStore, shortener URLShortener) *URLShortenerServer {
	return &URLShortenerServer{
		store:     store,
		shortener: shortener,
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
	expandedURL, found := u.store.Load(shortLink)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprint(w, expandedURL)
}

func (u *URLShortenerServer) processShortURL(w http.ResponseWriter, r *http.Request) {
	baseURL := strings.TrimPrefix(r.URL.Path, ShortenRoute)
	shortURL := u.shortener.ShortenBaseURL(baseURL)
	w.WriteHeader(http.StatusAccepted)
	u.store.Save(baseURL, shortURL)

}

type URLStore interface {
	Save(baseURL, shortLink string)
	Load(shortLink string) (string, bool)
}