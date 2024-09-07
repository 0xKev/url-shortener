package shortener

import (
	"fmt"
	"strings"
	"sync"

	"github.com/0xKev/url-shortener/internal/base62"
)

// max counter limit  3521614606207 base62 encoding -> zzzzzzz max length of 7

const (
	defaultDomain          = "example.com/"
	defaultURLSuffixLength = 7
	defaultURLCounterLimit = 3521614606207
	defaultURLCounter      = 500

	ErrCounterLimitReached  = "counter limit exceeded: "
	ErrEmptyURL             = "can't shorten empty url"
	ErrNoDomainURL          = "can't shorten url without a domain"
	ErrShortURLDoesNotExist = ""
)

type ExceedCounterError struct {
	CurrentCounter uint64
	MaxCounter     uint64
}

func (e ExceedCounterError) Error() string {
	return fmt.Sprintf(ErrCounterLimitReached+"current count %d, max count %d", e.CurrentCounter, e.MaxCounter)
}

type ShortURLNotFoundError struct {
	ShortURL string
}

func (s ShortURLNotFoundError) Error() string {
	return fmt.Sprintf("%v does not exist in store", s.ShortURL)
}

type InvalidURLError struct {
	ErrorMsg     string
	SubmittedURL string
}

func (i InvalidURLError) Error() string {
	return fmt.Sprintf("invalid url %s, %v", i.ErrorMsg, i.SubmittedURL)
}

type Config struct {
	domain          string
	urlSuffixLength uint64
	urlCounterLimit uint64
	urlCounter      uint64
}

func NewDefaultConfig() *Config {
	return &Config{
		domain:          defaultDomain,
		urlSuffixLength: defaultURLSuffixLength,
		urlCounterLimit: defaultURLCounterLimit,
		urlCounter:      defaultURLCounter,
	}
}

func (c *Config) URLSuffixLength() uint64 {
	return c.urlSuffixLength
}

func (c *Config) Domain() string {
	return c.domain
}

func (c *Config) URLCounterLimit() uint64 {
	return c.urlCounterLimit
}

func (c *Config) URLCounter() uint64 {
	return c.urlCounter
}

func (c *Config) SetDomain(domain string) {
	c.domain = domain
}

func (c *Config) SetURLCounter(counter uint64) {
	c.urlCounter = counter
}

type URLShortener struct {
	urlMap map[string]string
	Config *Config
	mu     sync.Mutex
}

func NewURLShortener(config *Config) *URLShortener {
	if config == nil {
		config = NewDefaultConfig()
	}
	return &URLShortener{
		urlMap: make(map[string]string),
		Config: config,
	}
}

func (u *URLShortener) ShortenURL(link string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	over, err := u.isOverCounterLimit()

	if over {
		return "", err
	}

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

func (u *URLShortener) isOverCounterLimit() (bool, error) {
	if u.Config.urlCounter >= u.Config.urlCounterLimit {
		return true, ExceedCounterError{
			CurrentCounter: u.Config.urlCounter,
			MaxCounter:     u.Config.urlCounterLimit,
		}
	}
	return false, nil
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
	return "", ShortURLNotFoundError{ShortURL: link}
}

func (u *URLShortener) validateURL(link string) error {
	if link == "" {
		return InvalidURLError{ErrEmptyURL, link}
	}

	if !strings.Contains(link, ".") {
		return InvalidURLError{ErrNoDomainURL, link}
	}

	return nil
}

func (u *URLShortener) generateShortSuffix() string {
	u.Config.urlCounter++
	generatedSuffix := base62.Encode(u.Config.urlCounter)
	return generatedSuffix
}
