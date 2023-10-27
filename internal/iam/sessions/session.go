package sessions

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/xid"
)

var (
	ErrSessionID     = errors.New("error generating UUID")
	ErrCreateSession = errors.New("error creating session")
	ErrNoSession     = errors.New("error querying session")
	ErrNoCookie      = errors.New("error getting cookie")
)

type Session struct {
	Name string
	// TODO -- expiry
	rwc *sql.DB
}

func New(ctx context.Context, name string, rwc *sql.DB) (s *Session, err error) {
	s = &Session{Name: name, rwc: rwc}
	err = s.rwc.PingContext(ctx)
	return
}

func (s *Session) Set(w http.ResponseWriter, r *http.Request, profile xid.ID) (session uuid.UUID, err error) {
	session, err = uuid.NewV7()
	if err != nil {
		return uuid.Nil, errors.Join(ErrSessionID, err)
	}
	_, err = s.rwc.ExecContext(r.Context(), "INSERT INTO sessions (id, profile) VALUES (?, ?)", session, profile)
	if err != nil {
		return uuid.Nil, errors.Join(ErrCreateSession, err)
	}
	s.setCookie(w, r, session.String(), 24*time.Hour)
	return
}

func (s *Session) Get(w http.ResponseWriter, r *http.Request) (profile xid.ID, err error) {
	c, err := s.cookie(r)
	if err != nil {
		return xid.NilID(), errors.Join(ErrNoCookie, err)
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return xid.NilID(), errors.Join(ErrSessionID, err)
	}
	err = s.rwc.QueryRowContext(r.Context(), "SELECT s.profile FROM sessions AS s WHERE s.id = ?", session).Scan(&profile)
	if err != nil {
		return xid.NilID(), errors.Join(ErrNoSession, err)
	}
	return
}

func (s *Session) Delete(w http.ResponseWriter, r *http.Request) (err error) {
	c, err := s.cookie(r)
	if err != nil {
		return errors.Join(ErrNoCookie, err)
	}
	session, err := uuid.FromString(c.Value)
	if err != nil {
		return errors.Join(ErrSessionID, err)
	}
	_, err = s.rwc.ExecContext(r.Context(), "DELETE FROM sessions AS s WHERE s.id = ?", session)
	if err != nil {
		return err
	}
	s.setCookie(w, r, "", -1)
	return
}

func (s Session) cookie(r *http.Request) (c *http.Cookie, err error) {
	var (
		name   = s.Name
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
		name   = s.Name
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
