package base62_test

//https://math.tools/calculator/base/10-62

import (
	"testing"

	"github.com/0xKev/url-shortener/internal/base62"
)

func TestEncodeBase62(t *testing.T) {
	t.Run("correctly encodes and return 7 char length", func(t *testing.T) {
		cases := []struct {
			original uint64
			encoded  string
		}{
			{0, "0000000"},
			{1, "0000001"},
			{5, "0000005"},
			{9, "0000009"},
			{10, "000000A"},
			{500, "0000084"},
			{3521614606207, "zzzzzzz"},
		}

		for _, c := range cases {
			got := base62.Encode(c.original)

			if got != c.encoded {
				t.Errorf("want %v but got %v", c.encoded, got)
			}
		}
	})
}
