package urlrenderer

import (
	"embed"
	"html/template"
	"io"

	"github.com/0xKev/url-shortener/internal/model"
)

//go:embed "templates/*.gohtml"
var urlPairTemplates embed.FS

//go:embed "static/css/output.css"
var static embed.FS

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

func GetStaticFS() embed.FS {
	return static
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

func (u *URLPairRenderer) RenderInvalidUserInput(w io.Writer, urlPair model.URLPair) error {
	// TODO(HIGH): add templ render for invalid user input (add an error message field to the urlPair)
	if err := u.templ.ExecuteTemplate(w, "invalid_user_input.gohtml", urlPair); err != nil {
		return err
	}
	return nil
}
