package http

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	httputil "github.com/adoublef/coffeesaurus/http"
	"github.com/adoublef/coffeesaurus/internal/iam"
	"github.com/adoublef/coffeesaurus/internal/iam/sqlite3"
	"github.com/go-chi/chi/v5"
	"github.com/superfly/litefs-go"
)

func (s *Service) handleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := s.a.Configs().Get(chi.URLParam(r, "provider"))
		if err != nil {
			httputil.Error(w, err, http.StatusBadRequest)
			return
		}
		ou, err := s.a.HandleCallback(w, r, c)
		if err != nil {
			httputil.Error(w, err, http.StatusUnauthorized)
			return
		}
		u := iam.NewUser(ou.ID, ou.Login, ou.Photo, ou.Name)
		// check profile exists
		// don't use `sql` errors inside this module
		// try to use domain level errors
		found, err := sqlite3.LookUpProfile(r.Context(), s.db.Raw(), ou.Login)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := litefs.WithHalt(s.db.String(), func() (err error) {
				return sqlite3.RegisterUser(context.Background(), s.db.Raw(), u)
			}); err != nil {
				httputil.Error(w, err, http.StatusInternalServerError)
				return
			}
		case err != nil:
			httputil.Error(w, err, http.StatusInternalServerError)
			return
		}
		if err := litefs.WithHalt(s.ss.String(), func() error {
			var id = u.ID()
			if found != nil {
				id = found.ID
			}
			_, err := s.ss.Set(w, r, id)
			return err
		}); err != nil {
			httputil.Error(w, err, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}
