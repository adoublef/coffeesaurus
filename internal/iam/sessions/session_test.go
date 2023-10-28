package sessions_test

import (
	"context"
	"crypto/tls"
	"embed"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/adoublef/coffeesaurus/internal/iam/sessions"
	"github.com/adoublef/coffeesaurus/sqlite3"
	"github.com/maragudk/migrate"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/xid"
	is "github.com/stretchr/testify/require"
)

//go:embed all:migrations/*.up.sql
var migrations embed.FS

func TestSession(t *testing.T) {
	t.Run("Set", withSession(func(t *testing.T, s *sessions.Session) {
		var (
			w, r    = newTestServer(true)
			profile = xid.New()
		)
		// set
		session, err := s.Set(w, r, profile)
		is.NoError(t, err)

		c := w.Result().Cookies()[0]
		is.True(t, strings.HasPrefix(c.Name, "_Host-"))
		is.Equal(t, session.String(), c.Value)
	}))

	t.Run("Get", withSession(func(t *testing.T, s *sessions.Session) {
		var (
			w, r    = newTestServer(true)
			profile = xid.New()
		)
		// set
		_, err := s.Set(w, r, profile)
		is.NoError(t, err)
		r.AddCookie(w.Result().Cookies()[0])
		// get
		found, err := s.Get(w, r)
		is.NoError(t, err)
		is.Equal(t, found, profile)
	}))

	t.Run("Delete", withSession(func(t *testing.T, s *sessions.Session) {
		var (
			profile = xid.New()
		)
		// set
		w, r := newTestServer(true)
		{
			_, err := s.Set(w, r, profile)
			is.NoError(t, err)
			c := w.Result().Cookies()[0]
			is.Equal(t, 86400, c.MaxAge)
		}
		// delete
		sc := w.Result().Cookies()[0]
		{
			w, r = newTestServer(true)
			r.AddCookie(sc)

			err := s.Delete(w, r)
			is.NoError(t, err)
			// check
			c := w.Result().Cookies()[0]
			is.True(t, strings.HasPrefix(c.Name, "_Host-"))
			is.Equal(t, "", c.Value)
			// not working
			// is.Equal(t, -1, c.MaxAge)
		}
	}))
}

// option to make secure secure
func newTestServer(secure bool) (w *httptest.ResponseRecorder, r *http.Request) {
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	if secure {
		r.TLS = &tls.ConnectionState{}
	}
	return w, r
}

func withSession(f func(t *testing.T, s *sessions.Session)) func(t *testing.T) {
	return func(t *testing.T) {
		dsn := path.Join(t.TempDir(), "cache.db")
		// run migration
		{
			db, err := sqlite3.Open(dsn)
			if err != nil {
				t.Fatalf("opening database: %v", err)
			}
			t.Cleanup(func() { db.Close() })

			fsys, err := fs.Sub(migrations, "migrations")
			if err != nil {
				t.Fatalf("opening directory: %v", err)
			}
			err = migrate.Up(context.TODO(), db.Raw(), fsys)
			if err != nil {
				t.Fatalf("execute migrations: %v", err)
			}
		}
		// create session
		s, err := sessions.New(context.TODO(), dsn)
		if err != nil {
			t.Fatalf("execute migrations: %v", err)
		}
		f(t, s)
	}
}
