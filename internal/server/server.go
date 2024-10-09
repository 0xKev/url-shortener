package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	urlrenderer "github.com/0xKev/url-shortener"
	"github.com/0xKev/url-shortener/internal/model"
)

const (
	ExpandRoute             = "/expand/"
	ShortenRoute            = "/shorten"
	JsonContentType         = "application/json"
	HtmxRequestContentType  = "application/x-www-form-urlencoded"
	HtmxResponseContentType = "text/html; charset=utf-8"

	APIVersion      = "v1"
	APIExpandRoute  = "/api/" + APIVersion + ExpandRoute
	APIShortenRoute = "/api/" + APIVersion + ShortenRoute

	HtmxExpandRoute  = ExpandRoute
	HtmxShortenRoute = ShortenRoute
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
	router.Handle(APIShortenRoute, http.HandlerFunc(server.shortenHandler))
	router.Handle(APIExpandRoute, http.HandlerFunc(server.expandHandler))

	router.Handle(HtmxShortenRoute, http.HandlerFunc(server.shortenHandler))
	router.Handle(HtmxExpandRoute, http.HandlerFunc(server.expandHandler))

	// log.Printf("Routes registered: /, %s, %s", ShortenRoute, ExpandRoute)

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
}

func (u *URLShortenerServer) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if u.isAPIRequest(r) {
		u.processAPIShortURL(w, r)
	} else if u.isHTMXRequest(r) {
		u.processHTMXShortURL(w, r)
	}
}

func (u *URLShortenerServer) expandHandler(w http.ResponseWriter, r *http.Request) {
	if u.isAPIRequest(r) {
		w.Header().Set("Content-Type", JsonContentType)
		u.showAPIExpandedURL(w, r)
	} else if u.isHTMXRequest(r) {
		w.Header().Set("Content-Type", HtmxResponseContentType)
		u.showHTMXExpandedURL(w, r)
	}

}

func (u *URLShortenerServer) isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (u *URLShortenerServer) isAPIRequest(r *http.Request) bool {
	return r.Header.Get("Content-Type") == JsonContentType
}

func (u *URLShortenerServer) showHTMXExpandedURL(w http.ResponseWriter, r *http.Request) {
	shortSuffix := strings.TrimPrefix(r.URL.Path, HtmxExpandRoute)
	baseURL, found := u.store.Load(shortSuffix)
	if !found {
		http.Error(w, "baseURL not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Redirect", baseURL)
}

func (u *URLShortenerServer) showAPIExpandedURL(w http.ResponseWriter, r *http.Request) {
	shortSuffix := strings.TrimPrefix(r.URL.Path, APIExpandRoute)

	expandedURL, found := u.store.Load(shortSuffix)
	w.Header().Set("Content-Type", JsonContentType)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "URL not found"})
		return
	}
	json.NewEncoder(w).Encode(u.getURLPair(shortSuffix, expandedURL))
}

func (u *URLShortenerServer) getURLPair(shortURL, baseURL string) model.URLPair {
	return model.URLPair{ShortSuffix: shortURL, BaseURL: baseURL}
}

func (u *URLShortenerServer) processAPIShortURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var urlPair *model.URLPair
	var err error

	if r.Header.Get("Content-Type") == JsonContentType {
		w.Header().Set("Content-Type", JsonContentType) // set response to Json contenttype if request same
		urlPair, err = u.processJSONShortURL(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Header.Get("HX-Request") == "true" {
		urlPair, err = u.processHTMXShortURL(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	u.store.Save(*urlPair)
}

func (u *URLShortenerServer) processHTMXShortURL(w http.ResponseWriter, r *http.Request) (*model.URLPair, error) {
	baseURL := r.FormValue("base-url")
	if baseURL == "" {
		return nil, errors.New("could not process baseURL from HTMX request")
	}
	shortSuffix, err := u.shortener.ShortenURL(baseURL)
	if err != nil {
		return nil, errors.New("could not shorten baseURL: " + err.Error())
	}

	urlPair := model.URLPair{BaseURL: baseURL, ShortSuffix: shortSuffix}
	err = u.renderer.Render(w, urlPair)
	if err != nil {
		panic(err)
	}
	return &urlPair, nil
}

func (u *URLShortenerServer) processJSONShortURL(w http.ResponseWriter, r *http.Request) (*model.URLPair, error) {
	var urlPair = model.URLPair{}
	err := json.NewDecoder(r.Body).Decode(&urlPair)

	if err != nil {
		return nil, errors.New("error decoding json")
	}

	// VALIDATE URL THEN RETURN ERROR IF INVALID
	if urlPair.BaseURL == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, errors.New("base url is empty")
	}

	shortSuffix, err := u.shortener.ShortenURL(urlPair.BaseURL)
	if err != nil {
		// issue might be here
		return nil, errors.New("could not shorten baseURL: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)

	urlPair.ShortSuffix = shortSuffix
	err = json.NewEncoder(w).Encode(urlPair)
	if err != nil {
		return nil, errors.New("could not encode to JSON: " + err.Error())
	}

	return &urlPair, nil
}

type URLStore interface {
	Save(model.URLPair) error
	Load(shortSuffix string) (string, bool)
}
