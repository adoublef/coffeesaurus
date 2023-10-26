package http

import (
	"net/http"

	"github.com/adoublef/coffeesaurus/env"
)

func (s *Service) handleIndex() http.HandlerFunc {
	region := env.Must("FLY_REGION")
	return func(w http.ResponseWriter, r *http.Request) {
		// NOTE static fonts and styles handled by external project
		t.ExecuteHTTP(w, r, "index", region)
	}
}
