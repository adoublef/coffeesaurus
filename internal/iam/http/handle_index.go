package http

import (
	"net/http"

	"github.com/adoublef/coffeesaurus/env"
)

func (s *Service) handleIndex() http.HandlerFunc {
	region := env.Must("FLY_REGION")
	return func(w http.ResponseWriter, r *http.Request) {
		if id, err := s.ss.Get(w, r); err != nil {
			// http.Error
			t.ExecuteHTTP(w, r, "index", region)
		} else {
			// get user by id
			t.ExecuteHTTP(w, r, "profile", id.String())
		}
	}
}
