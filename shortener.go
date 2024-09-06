package shortener

import (
	"fmt"
	"strings"

	"github.com/0xKev/url-shortener/internal/base62"
)

// max counter limit  3521614606207 base62 encoding -> zzzzzzz max length of 7

type Config struct {
	domain          string
	urlSuffixLength int
	urlCounterLimit int
}

func (c *Config) URLSuffixLength() int {
	return c.urlSuffixLength
}

func (c *Config) Domain() string {
	return c.domain
}

type URLShortener struct {
	urlMap     map[string]string
	urlCounter uint64
	Config     *Config
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urlMap:     make(map[string]string),
		urlCounter: 500,
		Config: &Config{
			domain:          "s.nykevin.com/",
			urlSuffixLength: 7,
			urlCounterLimit: 3521614606207,
		},
	}
}

func (u *URLShortener) ShortenURL(link string) (string, error) {
	if err := u.validateURL(link); err != nil {
		return "", err
	}

	shortLink, exists := u.urlMap[link]

	if exists {
		return shortLink, nil
	} else {
		u.urlMap[link] = fmt.Sprint(u.Config.domain + u.generateShortSuffix())
	}

	return u.urlMap[link], nil
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
	generatedSuffix := base62.Encode(u.urlCounter)
	u.urlCounter++
	return generatedSuffix
}
