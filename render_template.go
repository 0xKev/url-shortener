package urlrenderer

import (
	"embed"
	"html/template"
	"io"

	"github.com/0xKev/url-shortener/internal/model"
)

var (
	//go:embed "templates/*.gohtml"
	//go:embed "static/css/output.css"
	urlPairTemplates embed.FS
)

type URLPairRenderer struct {
	templ *template.Template
}

func NewURLPairRenderer() (*URLPairRenderer, error) {
	templ, err := template.ParseFS(urlPairTemplates, "templates/*.gohtml")
	if err != nil {
		return nil, err
	}

	return &URLPairRenderer{templ: templ}, nil
}

func (u *URLPairRenderer) Render(w io.Writer, urlPair model.URLPair) error {
	if err := u.templ.ExecuteTemplate(w, "url_pair.gohtml", urlPair); err != nil {
		return err
	}
	return nil
}

func (u *URLPairRenderer) RenderIndex(w io.Writer) error {
	if err := u.templ.ExecuteTemplate(w, "index.gohtml", nil); err != nil {
		return err
	}
	return nil
}
