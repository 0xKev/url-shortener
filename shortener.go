package shortener

import (
	"fmt"
	"strconv"
)

const Domain = "s.nykevin.com/"
const UrlSuffixLength = 7
const urlCounterLimit = 3521614606207 // base62 encoding -> zzzzzzz max length of 7

var (
	urlMap     = make(map[string]string)
	urlCounter = 500
)

func ShortenURL(link string) (string, error) {
	if link == "" {
		return "", fmt.Errorf("can't shorten invalid url %v", link)
	}

	shortLink, exists := urlMap[link]

	if exists {
		return shortLink, nil
	} else {
		urlMap[link] = fmt.Sprint(Domain + generateShortSuffix())
	}

	return urlMap[link], nil
}

func generateShortSuffix() string {

	generatedSuffix := "abcd" + strconv.FormatInt(int64(urlCounter), 10)
	urlCounter++
	return generatedSuffix
}
