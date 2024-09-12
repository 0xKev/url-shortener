package shortener

import (
	"fmt"
	"strings"
	"sync"
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
	Config  *Config
	mu      sync.Mutex
	encoder Encoder
}

func NewURLShortener(config *Config, encoder Encoder) *URLShortener {
	if config == nil {
		config = NewDefaultConfig()
	}
	return &URLShortener{
		Config:  config,
		encoder: encoder,
	}
}

func (u *URLShortener) ShortenURL(baseURL string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	over, err := u.isOverCounterLimit()

	if over {
		return "", err
	}

	if err := u.validateURL(baseURL); err != nil {
		return "", err
	}

	return u.generateShortSuffix(), nil
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

func (u *URLShortener) validateURL(link string) error {
	if link == "" {
		return InvalidURLError{ErrEmptyURL, link}
	}

	if !strings.Contains(link, ".") {
		return InvalidURLError{ErrNoDomainURL, link}
	}

	return nil
}

type Encoder interface {
	Encode(num uint64) string
}

func (u *URLShortener) generateShortSuffix() string {
	u.Config.urlCounter++
	generatedSuffix := u.encoder.Encode(u.Config.urlCounter)
	return generatedSuffix
}
