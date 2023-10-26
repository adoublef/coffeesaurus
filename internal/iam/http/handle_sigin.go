package http

import (
	"net/http"
	"time"

	"github.com/adoublef/coffeesaurus/env"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (s *Service) handleSignIn() http.HandlerFunc {
	p := oauth2.Config{
		ClientID:     env.Must("GOOGLE_CLIENT_ID"),
		ClientSecret: env.Must("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback/google",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		url , err := signIn(w, r, &p)
		if err !=nil{
			http.Error(w, "Failed to create auth code url", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(url))
	}
}

/* oauth2 */
func cookieName(r *http.Request, name string) string {
	if secure := r.TLS != nil; secure {
		name = "__Host-" + name
	}
	return name
}

func signIn(w http.ResponseWriter, r *http.Request, p *oauth2.Config) (redirectURL string, err error) {
	s, err := randomState(16)
	if err != nil {
		return "", err
	}
	redirectURL = p.AuthCodeURL(s)

	c := http.Cookie{
		// 10 minutes
		MaxAge:   int((10 * time.Minute).Seconds()),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Value:    s,
		Name:     cookieName(r, site),
	}
	http.SetCookie(w, &c)
	return
}

func randomState(n int) (state string, err error) {
	return
}

const (
	oauth = "oauth-session"
	site  = "site-session"
)
