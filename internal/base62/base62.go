package base62

import "strings"

const base62Digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Encode(num uint64) string {
	if num == 0 {
		return "0"
	}
	var base62 []string

	for num > 0 {
		remainder := num % 62
		base62 = append(base62, string(base62Digits[remainder]))
		num /= 62
	}

	return strings.Join(base62, "")
}
