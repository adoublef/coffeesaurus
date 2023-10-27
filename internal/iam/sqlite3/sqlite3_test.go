package sqlite3_test

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/adoublef/coffeesaurus/internal/iam"
	"github.com/adoublef/coffeesaurus/internal/iam/sqlite3"
	"github.com/adoublef/coffeesaurus/oauth2"
	"github.com/maragudk/migrate"
	_ "github.com/mattn/go-sqlite3"
	is "github.com/stretchr/testify/require"
)

//go:embed all:migrations/*.up.sql
var migrations embed.FS

func TestSqlite3(t *testing.T) {
	// Example Errors
	// 	- FOREIGN KEY constraint failed
	// 	- UNIQUE constraint failed: profiles.login
	t.Run("RegisterUser", withClient(func(t *testing.T, db *sql.DB) {
		// create a new user
		a := iam.NewUser(oauth2.NewID("google", "1"),
			"alpha@gmail.com", "https://avatar.google.com/alpha", "Alpha")

		err := sqlite3.RegisterUser(context.Background(), db, a)
		is.NoError(t, err, "register 'a")

		// fail to add duplicate user
		err = sqlite3.RegisterUser(context.Background(), db, a)
		is.Error(t, err, "duplicate 'a'")

		// add second user
		b := iam.NewUser(oauth2.NewID("google", "2"),
			"bravo@gmail.com", "https://avatar.google.com/bravo", "Bravo")

		err = sqlite3.RegisterUser(context.Background(), db, b)
		is.NoError(t, err, "register 'b'")
	}))

	// Lookup user `login`
	t.Run("LookUpProfile", withClient(func(t *testing.T, db *sql.DB) {
		// create a new user
		a := iam.NewUser(oauth2.NewID("google", "1"),
			"alpha@gmail.com", "https://avatar.google.com/alpha", "Alpha")
		err := sqlite3.RegisterUser(context.Background(), db, a)
		is.NoError(t, err, "register 'a")

		// lookup profile
		found, err := sqlite3.LookUpProfile(context.Background(), db, "alpha@gmail.com")
		is.NoError(t, err, "lookup 'a'")
		is.Equal(t, a.Profile, found)
	}))
}

func withClient(f func(t *testing.T, db *sql.DB)) func(t *testing.T) {
	args := []string{"_journal=wal", "_timeout=5000", "_synchronous=normal", "_fk=true"}
	return func(t *testing.T) {
		dsn := path.Join(t.TempDir(), "test.db")
		db, err := sql.Open("sqlite3", dsn+"?"+strings.Join(args, "&"))
		if err != nil {
			t.Fatalf("opening database: %s", err)
		}
		// close file
		t.Cleanup(func() { db.Close() })
		// run migration
		fsys, err := fs.Sub(migrations, "migrations")
		if err != nil {
			t.Fatalf("opening migrations directory: %s", err)
		}
		err = migrate.Up(context.TODO(), db, fsys)
		if err != nil {
			t.Fatalf("execute migration files: %s", err)
		}
		f(t, db)
	}
}
