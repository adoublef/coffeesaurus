package http

import (
	"embed"
	"net/http"

	"github.com/adoublef-go/template"
	"github.com/go-chi/chi/v5"
)

//go:embed all:*.html
var fsys embed.FS

// change API to allow any pattern
var t = template.Must(fsys, template.Partials(false))

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService() *Service {
	s := Service{
		m: chi.NewMux(),
		// db for ping
	}
	s.routes()
	return &s
}

func (s *Service) routes() {
	// simple index page
	s.m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// NOTE static fonts and styles handled by external project
		t.ExecuteHTTP(w, r, "index", nil)
	})
}
