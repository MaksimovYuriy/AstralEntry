package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
)

type Repo struct {
	db *sql.DB
}

var _ repo.UserRepo = (*Repo)(nil)

func New(db *sql.DB) repo.UserRepo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, user entity.User) error {
	const query = `
		INSERT INTO users (id, login, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (login) DO NOTHING
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return repo.ErrAlreadyExists
	}

	return nil
}

func (r *Repo) FindByLogin(ctx context.Context, login string) (entity.User, error) {
	const query = `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`

	var user entity.User
	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, repo.ErrNotFound
	}

	return user, err
}
