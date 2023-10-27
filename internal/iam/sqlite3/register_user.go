package sqlite3

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adoublef/coffeesaurus/internal/iam"
)

func RegisterUser(ctx context.Context, db *sql.DB, u *iam.User) (err error) {
	// begin
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("tx begin error: %w", err)
	}
	defer tx.Rollback()

	// try to insert profile
	_, err = tx.ExecContext(ctx, `
INSERT INTO profiles (id, login, photo_url, name)
VALUES (?, ?, ?, ?)
	`, u.Profile.ID, u.Profile.Login, u.Profile.Photo, u.Profile.Name)
	if err != nil {
		return err
	}

	// insert credentials
	_, err = tx.ExecContext(ctx, `
INSERT INTO credentials (oauth, profile)
VALUES (?, ?)
	`, u.OAuth2, u.Profile.ID)
	if err != nil {
		return err
	}
	// commit
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx commit error: %w", err)
	}

	return
}
