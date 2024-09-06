package base62

import (
	"slices"
	"strings"
)

const (
	base62Digits  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	encodedLength = 7
)

func Encode(num uint64) string {
	if num == 0 {
		return strings.Repeat("0", encodedLength)
	}
	var base62 []string

	for num > 0 {
		remainder := num % 62
		base62 = append(base62, string(base62Digits[remainder]))
		num /= 62
	}

	for len(base62) != encodedLength {
		base62 = append(base62, "0")
	}

	slices.Reverse(base62)

	return strings.Join(base62, "")
}
