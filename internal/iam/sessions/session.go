package sessions

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/adoublef/coffeesaurus/sqlite3"
	"github.com/gofrs/uuid"
	"github.com/rs/xid"
)

var (
	//go:embed all:migrations/*.up.sql
	migrations embed.FS
)

// Up will run through the migration files
func Up(ctx context.Context, dsn string) (err error) {
	db, err := sqlite3.Open(dsn)
	if err != nil {
		return fmt.Errorf("opening connection: %w", err)
	}
	defer db.Close()

	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("return file system: %w", err)
	}

	err = db.Up(ctx, fsys)
	if err != nil {
		return fmt.Errorf("run migration files: %w", err)
	}

	return nil
}

var (
	ErrSessionID     = errors.New("error generating UUID")
	ErrCreateSession = errors.New("error creating session")
	ErrNoSession     = errors.New("error querying session")
	ErrNoCookie      = errors.New("error getting cookie")
)

var (
	_ fmt.Stringer = (Session{})
	_ io.Closer    = (*Session)(nil)
)

type Session struct {
	// TODO -- expiry
	rwc *sqlite3.DB
}

// String implements fmt.Stringer.
func (s Session) String() string { return s.rwc.String() }

// Close implements io.Closer.
func (s *Session) Close() error { return s.rwc.Close() }

func New(ctx context.Context, dsn string) (s *Session, err error) {
	rwc, err := sqlite3.Open(dsn)
	if err != nil {
		return nil, err
	}
	s = &Session{rwc: rwc}
	err = s.rwc.Raw().PingContext(ctx)
	return
}

func (s Session) Set(w http.ResponseWriter, r *http.Request, profile xid.ID) (session uuid.UUID, err error) {
	session, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil, errors.Join(ErrSessionID, err)
	}
	_, err = s.rwc.Raw().ExecContext(r.Context(), "INSERT INTO sessions (id, profile) VALUES (?, ?)", session, profile)
	if err != nil {
		return uuid.Nil, errors.Join(ErrCreateSession, err)
	}
	s.setCookie(w, r, session.String(), 24*time.Hour)
	return
}

func (s Session) Get(w http.ResponseWriter, r *http.Request) (profile xid.ID, err error) {
	c, err := s.cookie(r)
	if err != nil {
		return xid.NilID(), errors.Join(ErrNoCookie, err)
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return xid.NilID(), errors.Join(ErrSessionID, err)
	}
	err = s.rwc.Raw().QueryRowContext(r.Context(), "SELECT s.profile FROM sessions AS s WHERE s.id = ?", session).Scan(&profile)
	if err != nil {
		return xid.NilID(), errors.Join(ErrNoSession, err)
	}
	return
}

func (s Session) Delete(w http.ResponseWriter, r *http.Request) (err error) {
	c, err := s.cookie(r)
	if err != nil {
		return errors.Join(ErrNoCookie, err)
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return errors.Join(ErrSessionID, err)
	}
	_, err = s.rwc.Raw().ExecContext(r.Context(), "DELETE FROM sessions AS s WHERE s.id = ?", session)
	if err != nil {
		return err
	}
	s.setCookie(w, r, "", -1)
	return
}

func (s Session) cookie(r *http.Request) (c *http.Cookie, err error) {
	var (
		name   = "site-session"
		secure = false
	)

	if secure = r.TLS != nil; secure {
		name = "_Host-" + name
	}
	c, err = r.Cookie(name)
	return
}

func (s Session) setCookie(w http.ResponseWriter, r *http.Request, value string, maxAge time.Duration) {
	var (
		name   = "site-session"
		secure = false
	)

	if secure = r.TLS != nil; secure {
		name = "_Host-" + name
	}
	c := http.Cookie{
		Name:     name,
		Secure:   secure,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int(maxAge.Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &c)
}
