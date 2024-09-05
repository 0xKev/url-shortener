package base62_test

//https://math.tools/calculator/base/10-62

import (
	"testing"

	"github.com/0xKev/url-shortener/internal/base62"
)

func TestEncodeBase62(t *testing.T) {
	cases := []struct {
		original uint64
		encoded  string
	}{
		{0, "0"},
		{1, "1"},
		{5, "5"},
		{9, "9"},
		{10, "A"},
		{3521614606207, "zzzzzzz"},
	}

	for _, c := range cases {
		got := base62.Encode(c.original)

		if got != c.encoded {
			t.Errorf("want %v but got %v", c.encoded, got)
		}
	}
}
