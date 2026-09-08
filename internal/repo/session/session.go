package session

import (
	"context"
	"database/sql"

	"github.com/maksimovyuriy/astralentry/internal/entity"
	"github.com/maksimovyuriy/astralentry/internal/repo"
)

type Repo struct {
	db *sql.DB
}

var _ repo.SessionRepo = (*Repo)(nil)

func New(db *sql.DB) repo.SessionRepo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, session entity.Session) error {
	const query = `
		INSERT INTO sessions (token_hash, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		session.TokenHash,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
	)
	return err
}

func (r *Repo) FindByTokenHash(ctx context.Context, tokenHash []byte) (entity.Session, error) {
	const query = `
		SELECT token_hash, user_id, created_at, expires_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP
	`

	var session entity.Session
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.TokenHash,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
	)
	return session, err
}

func (r *Repo) DeleteByTokenHash(ctx context.Context, tokenHash []byte) error {
	const query = `DELETE FROM sessions WHERE token_hash = $1`

	_, err := r.db.ExecContext(ctx, query, tokenHash)
	return err
}
