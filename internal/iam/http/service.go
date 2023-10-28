package http

import (
	"embed"
	"net/http"

	"github.com/adoublef-go/template"
	"github.com/adoublef/coffeesaurus/internal/iam/oauth2"
	"github.com/adoublef/coffeesaurus/internal/iam/oauth2/google"
	"github.com/adoublef/coffeesaurus/internal/iam/sessions"
	"github.com/adoublef/coffeesaurus/sqlite3"
	"github.com/go-chi/chi/v5"
)

//go:embed all:*.html
var fsys embed.FS

// change API to allow any pattern
var t = template.Must(fsys, template.Partials(false))

var _ http.Handler = (*Service)(nil)

type Service struct {
	m *chi.Mux
	a *oauth2.Authenticator
	// this should be passed through middleware
	ss *sessions.Session
	db *sqlite3.DB
}

// ServeHTTP implements http.Handler.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func NewService(db *sqlite3.DB, ss *sessions.Session) *Service {
	s := Service{
		m:  chi.NewMux(),
		a:  &oauth2.Authenticator{},
		db: db,
		ss: ss,
	}
	s.routes()
	return &s
}

func (s *Service) routes() {
	baseURL := "http://localhost:8080"

	ggURL := oauth2.RedirectURL(baseURL + "/callback/google")
	s.a.Configs().Set("google", google.NewConfig(ggURL))

	s.m.Get("/", s.handleIndex())
	s.m.Get("/signin/{provider}", s.handleSignIn())
	s.m.Get("/callback/{provider}", s.handleCallback())
	s.m.Get("/signout", s.handleSignOut())
}
