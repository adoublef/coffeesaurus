package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/maragudk/migrate"
	_ "github.com/mattn/go-sqlite3"
)

var (
	_ fmt.Stringer = DB{}
	_ io.Closer    = (*DB)(nil)
)

type DB struct {
	rwc *sql.DB
	dsn string
}

// Close implements io.Closer.
func (d *DB) Close() error { return d.rwc.Close() }

// String implements fmt.Stringer.
func (d DB) String() string { return d.Name() }

// Name returns the database filename
func (d DB) Name() string { return d.dsn }

// Raw returns the sql.DB type
func (d DB) Raw() *sql.DB { return d.rwc }

// Up runs the migration files inside an fs.FS object
func (d *DB) Up(ctx context.Context, fsys fs.FS) (err error) {
	err = migrate.Up(ctx, d.rwc, fsys)
	return
}

func Open(dsn string) (*DB, error) {
	args := []string{"_journal=wal", "_timeout=5000", "_synchronous=normal", "_fk=true"}
	rwc, err := sql.Open("sqlite3", dsn+"?"+strings.Join(args, "&"))
	if err != nil {
		return nil, err
	}

	return &DB{rwc: rwc, dsn: dsn}, nil
}
