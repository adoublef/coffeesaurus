package sqlite3

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/maragudk/migrate"
)

var (
	//go:embed all:migrations/*.up.sql
	migrations embed.FS
)

func Up(ctx context.Context, dsn string) (err error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("opening connection: %w", err)
	}
	defer db.Close()

	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("return file system: %w", err)
	}

	err = migrate.Up(ctx, db, fsys)
	if err != nil {
		return fmt.Errorf("run migration files: %w", err)
	}

	return nil
}

func Ping(ctx context.Context, db *sql.DB, tableName string) (err error) {
	var n int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&n)
	if err != nil {
		return fmt.Errorf("ping database for %s table: %w", tableName, err)
	}

	return nil
}
