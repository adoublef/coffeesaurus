package sqlite3

import (
	"context"
	"database/sql"

	"github.com/adoublef/coffeesaurus/internal/iam"
)

func LookUpProfile(ctx context.Context, db *sql.DB, login string) (*iam.Profile, error) {
	var p iam.Profile
	err := db.QueryRowContext(ctx, `
SELECT p.id, p.login, p.photo_url, p.name
FROM profiles AS p
WHERE p.login = ?
LIMIT 1
	`, login).Scan(&p.ID, &p.Login, &p.Photo, &p.Name)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
