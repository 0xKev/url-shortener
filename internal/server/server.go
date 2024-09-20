package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	ExpandRoute     = "/expand/"
	ShortenRoute    = "/shorten/"
	JsonContentType = "application/json"
)

type URLShortener interface {
	ShortenURL(baseURL string) (string, error)
}

type URLPair struct {
	ShortURL string
	BaseURL  string
}

type URLShortenerServer struct {
	store     URLStore
	shortener URLShortener
	http.Handler
}

func NewURLShortenerServer(store URLStore, shortener URLShortener) *URLShortenerServer {
	server := &URLShortenerServer{
		store:     store,
		shortener: shortener,
	}

	router := http.NewServeMux()
	router.Handle(ShortenRoute, http.HandlerFunc(server.shortenHandler))
	router.Handle(ExpandRoute, http.HandlerFunc(server.expandHandler))

	server.Handler = router

	return server

}

func (u *URLShortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	u.processShortURL(w, r)
}

func (u *URLShortenerServer) expandHandler(w http.ResponseWriter, r *http.Request) {
	u.showExpandedURL(w, r)
}

func (u *URLShortenerServer) showExpandedURL(w http.ResponseWriter, r *http.Request) {
	shortLink := strings.TrimPrefix(r.URL.Path, ExpandRoute)
	expandedURL, found := u.store.Load(shortLink)
	w.Header().Set("content-type", JsonContentType)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(u.getURLPair(shortLink, expandedURL))
	fmt.Fprint(w, expandedURL)
}

func (u *URLShortenerServer) getURLPair(shortURL, baseURL string) URLPair {
	return URLPair{ShortURL: shortURL, BaseURL: baseURL}
}

func (u *URLShortenerServer) processShortURL(w http.ResponseWriter, r *http.Request) {
	baseURL := strings.TrimPrefix(r.URL.Path, ShortenRoute)
	shortURL, err := u.shortener.ShortenURL(baseURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not shorten URL: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", JsonContentType)
	w.WriteHeader(http.StatusOK)
	response := u.getURLPair(shortURL, baseURL)
	json.NewEncoder(w).Encode(response)
	fmt.Fprint(w, shortURL)
	u.store.Save(shortURL, baseURL)
}

type URLStore interface {
	Save(shortLink, baseURL string) error
	Load(shortLink string) (string, bool)
}
