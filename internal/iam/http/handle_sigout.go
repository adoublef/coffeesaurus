package http

import (
	"net/http"
)

func (s *Service) handleSignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = s.ss.Delete(w, r)
		// go home
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
