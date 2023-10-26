package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.a.Configs().Get(chi.URLParam(r, "provider"))
		if err != nil {
			http.Error(w, "no provider exists", http.StatusBadRequest)
			return
		}

		url, err := s.a.SignIn(w, r, c)
		if err != nil {
			http.Error(w, "Failed to create auth code url", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}
}
