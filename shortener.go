package shortener

import (
	"fmt"
	"strings"
	"sync"

	"github.com/0xKev/url-shortener/internal/base62"
)

// max counter limit  3521614606207 base62 encoding -> zzzzzzz max length of 7

type Config struct {
	domain          string
	urlSuffixLength int
	urlCounterLimit int
	urlCounter      uint64
	mu              sync.Mutex
}

func (c *Config) URLSuffixLength() int {
	return c.urlSuffixLength
}

func (c *Config) Domain() string {
	return c.domain
}

type URLShortener struct {
	urlMap map[string]string
	Config *Config
}

func NewURLShortener(startCounter uint64) *URLShortener {
	return &URLShortener{
		urlMap: make(map[string]string),
		Config: &Config{
			domain:          "s.nykevin.com/",
			urlSuffixLength: 7,
			urlCounterLimit: 3521614606207,
			urlCounter:      500,
		},
	}
}

func (u *URLShortener) ShortenURL(link string) (string, error) {
	u.Config.mu.Lock()
	defer u.Config.mu.Unlock()
	if err := u.validateURL(link); err != nil {
		return "", err
	}

	shortLink, exists := u.urlMap[link]

	if exists {
		return shortLink, nil
	} else {
		u.urlMap[link] = u.generateShortURL()
	}

	return u.urlMap[link], nil
}

func (u *URLShortener) generateShortURL() string {
	return fmt.Sprint(u.Config.domain + u.generateShortSuffix())
}

func (u *URLShortener) ExpandURL(link string) (string, error) {
	for originalURL, shortURL := range u.urlMap {
		if shortURL == link {
			return originalURL, nil
		}
	}
	return "", fmt.Errorf("%v does not exist in store", link)
}

func (u *URLShortener) validateURL(link string) error {
	if link == "" {
		return fmt.Errorf("can't shorten empty url %v", link)
	}

	if !strings.Contains(link, ".") {
		return fmt.Errorf("can't shorten url without a domain %v", link)
	}

	return nil
}

func (u *URLShortener) generateShortSuffix() string {
	generatedSuffix := base62.Encode(u.Config.urlCounter)
	u.Config.urlCounter++
	return generatedSuffix
}
