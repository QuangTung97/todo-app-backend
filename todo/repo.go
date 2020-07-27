package todo

import (
	"context"
	"time"

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

func (repo *repository) transact(ctx context.Context, f func(tx *sqlx.Tx) error) error {
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return newErr
		}
		return err
	}

	return tx.Commit()
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

func (repo *repository) saveTodoList(ctx context.Context) todoListSaver {
	return func(accountID int, name string) (int, time.Time, error) {
		now := time.Now()
		query := repo.db.Rebind(`
            INSERT INTO todo_list (
                name, account_id,
                created_at, updated_at)
            VALUES (?, ?, ?, ?)
            `)

		res, err := repo.db.ExecContext(ctx, query, name, accountID, now, now)
		if err != nil {
			return 0, now, err
		}

		id, err := res.LastInsertId()
		return int(id), now, err
	}
}

func (repo *repository) getTodoList(ctx context.Context, tx *sqlx.Tx) todoListGetter {
	return func(id int) (todoList, error) {
		type Result struct {
			ID        int       `db:"id"`
			Name      string    `db:"name"`
			AccountID int       `db:"account_id"`
			CreatedAt time.Time `db:"created_at"`
			UpdatedAt time.Time `db:"updated_at"`
		}
		r := Result{}

		query := repo.db.Rebind(`
            SELECT id, name, account_id,
                created_at, updated_at
            FROM todo_list WHERE id = ?
            `)

		err := tx.GetContext(ctx, &r, query, id)
		return todoList{
			id:        r.ID,
			accountID: r.AccountID,
			name:      r.Name,
			createdAt: r.CreatedAt,
			updatedAt: r.UpdatedAt,
		}, err
	}
}

func (repo *repository) updateTodoList(ctx context.Context, tx *sqlx.Tx) todoListUpdater {
	return func(id int, name string) (time.Time, error) {
		now := time.Now()
		query := repo.db.Rebind(`
            UPDATE todo_list SET name = ?, updated_at = ? WHERE id = ?`)
		_, err := tx.ExecContext(ctx, query, name, now, id)
		return now, err
	}
}

func (repo *repository) getTodoListsByAccount(ctx context.Context) todoListsByAccountGetter {
	return func(accountID int) ([]todoList, error) {
		type Todo struct {
			ID        int       `db:"id"`
			AccountID int       `db:"account_id"`
			Name      string    `db:"name"`
			CreatedAt time.Time `db:"created_at"`
			UpdatedAt time.Time `db:"updated_at"`
		}

		todos := make([]Todo, 0)
		result := make([]todoList, 0)

		query := repo.db.Rebind(`
            SELECT id, account_id, name, created_at, updated_at
            FROM todo_list WHERE account_id = ?`)

		err := repo.db.SelectContext(ctx, &todos, query, accountID)
		if err != nil {
			return result, err
		}

		for _, t := range todos {
			result = append(result, todoList{
				id:        t.ID,
				accountID: t.AccountID,
				name:      t.Name,
				createdAt: t.CreatedAt,
				updatedAt: t.UpdatedAt,
			})
		}

		return result, nil
	}
}

func (repo *repository) deleteTodoList(ctx context.Context, tx *sqlx.Tx) todoListDeleter {
	return func(id int) error {
		query := repo.db.Rebind(`
            DELETE FROM todo_list WHERE id = ?`)
		_, err := tx.ExecContext(ctx, query, id)
		return err
	}
}
