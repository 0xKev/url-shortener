package urlrenderer_test

import (
	"bytes"
	"io"
	"testing"

	urlrenderer "github.com/0xKev/url-shortener"
	"github.com/0xKev/url-shortener/internal/model"
	approvals "github.com/approvals/go-approval-tests"
)

func TestRender(t *testing.T) {
	var (
		urlPair = model.URLPair{
			ShortSuffix: "0000001",
			BaseURL:     "google.com",
			Domain:      "https://s.domain.com/",
		}
	)

	urlPairRenderer, err := urlrenderer.NewURLPairRenderer()

	if err != nil {
		t.Fatal(err)
	}
	t.Run("converts a single urlPair into HTML", func(t *testing.T) {
		buf := bytes.Buffer{}

		if err := urlPairRenderer.Render(&buf, urlPair); err != nil {
			t.Fatal(err)
		}

		approvals.VerifyString(t, buf.String())
	})

	//  make sure to test index later same url pair
	t.Run("renders index.gohtml correctly", func(t *testing.T) {
		buf := bytes.Buffer{}

		if err := urlPairRenderer.RenderIndex(&buf); err != nil {
			t.Fatal(err)
		}

		approvals.VerifyString(t, buf.String())
	})

	t.Run("renders invalid_user_input.gohtml with error message when user submits invalid url", func(t *testing.T) {
		// TODO(HIGH): Add correct error page rendering
		buf := bytes.Buffer{}
		urlPair := model.URLPair{BaseURL: "bad-base-url", Error: "input link is not valid."}

		if err := urlPairRenderer.RenderInvalidUserInput(&buf, urlPair); err != nil {
			t.Fatal(err)
		}
		approvals.VerifyString(t, buf.String())
	})
}

func BenchmarkRender(b *testing.B) {
	var (
		urlPair = model.URLPair{
			ShortSuffix: "0000001",
			BaseURL:     "google.com",
		}
	)

	urlPairRenderer, err := urlrenderer.NewURLPairRenderer()

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		urlPairRenderer.Render(io.Discard, urlPair)
	}

}
