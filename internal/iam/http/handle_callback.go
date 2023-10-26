package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.a.Configs().Get(chi.URLParam(r, "provider"))
		if err != nil {
			http.Error(w, "no provider exists", http.StatusBadRequest)
			return
		}
		session, u, err := s.a.HandleCallback(w, r, c)
		if err != nil {
			http.Error(w, "Failed to authenticate", http.StatusUnauthorized)
			return
		}
		{
			newLine := "\n"
			fmt.Print(u.ID)
			fmt.Print(newLine)
			fmt.Print(u.Login)
			fmt.Print(newLine)
			fmt.Print(u.Photo)
			fmt.Print(newLine)
			fmt.Print(session)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
