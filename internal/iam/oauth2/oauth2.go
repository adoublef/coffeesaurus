package oauth2

import (
	"context"
	"crypto/subtle"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	ProviderGithub = "github"
	ProviderGoogle = "google"
)

var _ fmt.Stringer = (ID{})
var _ driver.Valuer = (*ID)(nil)

type ID struct {
	// Provider is the name of the oauth2 service
	Provider string
	// UserID is the id returned by the oauth2 service
	UserID string
}

// Value implements driver.Valuer.
func (i ID) Value() (driver.Value, error) {
	return i.String(), nil
}

func NewID(provider, value string) ID {
	return ID{Provider: provider, UserID: value}
}

// String implements Stringer
func (i ID) String() string { return i.Provider + "|" + i.UserID }

type UserInfo struct {
	// ID is a compound of the auth provider and the associated id
	ID    ID     `json:"id"`
	Photo string `json:"photo"`
	Login string `json:"login"`
	Name  string `json:"name"`
}

type Authenticator struct {
	cs Configs
}

// SignIn
func (a Authenticator) SignIn(w http.ResponseWriter, r *http.Request, cfg Config) (url string, err error) {
	state := oauth2.GenerateVerifier()
	c := http.Cookie{
		Name:     cookieName(r, oauth),
		Secure:   r.TLS != nil,
		Path:     "/",
		HttpOnly: true,
		Value:    state,
		// 10 minutes
		MaxAge:   int((10 * time.Minute).Seconds()),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &c)
	return cfg.AuthCodeURL(state), nil
}

// HandleCallback
func (a Authenticator) HandleCallback(w http.ResponseWriter, r *http.Request, p Config) (u *UserInfo, err error) {
	// get cookie
	cookie, err := getCookie(r, oauth)
	if err != nil {
		return nil, err
	}
	// compare with url of state on request
	if !compare(cookie.Value, r.FormValue("state")) {
		return nil, errors.New("state value mismatch")
	}
	cookie.Value = ""
	cookie.MaxAge = -1
	http.SetCookie(w, cookie) // set cookie
	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, httpClient)
	// exchange `code` for `tok`
	tok, err := p.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		return nil, fmt.Errorf("exchanging for token: %w", err)
	}
	// get `userinfo`
	u, err = p.UserInfo(ctx, tok)
	return
}

func (a *Authenticator) Configs() Configs {
	if a.cs == nil {
		a.cs = make(Configs)
	}
	return a.cs
}

func RedirectURL(url string) ConfigOption {
	return func(c *oauth2.Config) {
		c.RedirectURL = url
	}
}

type ConfigOption func(*oauth2.Config)

type Config interface {
	UserInfo(ctx context.Context, tok *oauth2.Token) (*UserInfo, error)
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type Configs map[string]Config

func (pp Configs) Get(key string) (Config, error) {
	p, ok := pp[key]
	if !ok {
		return nil, errors.New("provider not found")
	}

	return p, nil
}

func (pp Configs) Set(key string, p Config) {
	if _, found := pp[key]; !found {
		pp[key] = p
	}
}

// ----------------------------------------------------
func cookieName(r *http.Request, name string) string {
	if secure := r.TLS != nil; secure {
		name = "__Host-" + name
	}
	return name
}

func getCookie(r *http.Request, name string) (*http.Cookie, error) {
	name = cookieName(r, name)
	return r.Cookie(name)
}

func compare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) != 0
}

const (
	oauth = "oauth-session"
)