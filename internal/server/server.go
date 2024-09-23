package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	urlrenderer "github.com/0xKev/url-shortener"
	"github.com/0xKev/url-shortener/internal/model"
)

const (
	ExpandRoute     = "/expand/"
	ShortenRoute    = "/shorten/"
	JsonContentType = "application/json"
)

type URLShortener interface {
	ShortenURL(baseURL string) (string, error)
}

type URLShortenerServer struct {
	store     URLStore
	shortener URLShortener
	renderer  *urlrenderer.URLPairRenderer
	http.Handler
}

func NewURLShortenerServer(store URLStore, shortener URLShortener) *URLShortenerServer {
	renderer, err := urlrenderer.NewURLPairRenderer()

	if err != nil {
		panic(err) // panic only temp - handle err later
	}

	server := &URLShortenerServer{
		store:     store,
		shortener: shortener,
		renderer:  renderer,
	}

	router := http.NewServeMux()
	router.HandleFunc("/", server.indexHandler)
	router.Handle(ShortenRoute, http.HandlerFunc(server.shortenHandler))
	router.Handle(ExpandRoute, http.HandlerFunc(server.expandHandler))

	log.Printf("Routes registered: /, %s, %s", ShortenRoute, ExpandRoute)

	server.Handler = router

	return server

}

func (u *URLShortenerServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	if strings.TrimPrefix(r.URL.Path, "/") != "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	err := u.renderer.RenderIndex(w)

	if err != nil {
		panic(err)
	}
	log.Println("Index rendered successfully")
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
}

func (u *URLShortenerServer) getURLPair(shortURL, baseURL string) model.URLPair {
	return model.URLPair{ShortSuffix: shortURL, BaseURL: baseURL}
}

func (u *URLShortenerServer) processShortURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not shorten URL: %v", err), http.StatusInternalServerError)
		return
	}
	baseURL := string(body)

	// baseURL := strings.TrimPrefix(r.URL.Path, ShortenRoute)
	shortURL, err := u.shortener.ShortenURL(baseURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not shorten URL: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", JsonContentType)
	w.WriteHeader(http.StatusOK)
	response := u.getURLPair(shortURL, baseURL)
	json.NewEncoder(w).Encode(response)
	u.store.Save(shortURL, baseURL)
}

type URLStore interface {
	Save(shortLink, baseURL string) error
	Load(shortLink string) (string, bool)
}
