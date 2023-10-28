package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/adoublef/coffeesaurus/sqlite3"
	_ "github.com/mattn/go-sqlite3"
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

// Ping returns an error if a table does not exist
func Ping(ctx context.Context, db *sql.DB, tableName string) (err error) {
	var n int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&n)
	if err != nil {
		return fmt.Errorf("ping database for %s table: %w", tableName, err)
	}

	return nil
}
