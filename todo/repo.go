package todo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func newRepository(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (repo *repository) saveAccount(ctx context.Context) accountSaver {
	return func(username, hash string) error {
		query := repo.db.Rebind(
			`INSERT INTO account(username, password_hash) VALUES (?, ?)`)
		_, err := repo.db.ExecContext(ctx, query, username, hash)
		// TODO: unique contraint
		return err
	}
}
