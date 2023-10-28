package http

import (
	"net/http"

	"github.com/superfly/litefs-go"
)

func (s *Service) handleSignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		litefs.WithHalt(s.ss.String(),
			func() (err error) { return s.ss.Delete(w, r) })
		// go home
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
